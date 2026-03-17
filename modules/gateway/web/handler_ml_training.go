package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/plugin"
	"github.com/vistone/fingerprint/modules/profiles"
	tlsmod "github.com/vistone/fingerprint/modules/tls"
)

func (h *Handler) handleMLServiceTrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		http.Error(w, "MLService not enabled", http.StatusServiceUnavailable)
		return
	}

	h.trainingMu.Lock()
	if h.trainingActive {
		h.trainingMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Training already in progress (" + h.trainingPhase + ")",
		})
		return
	}
	h.trainingActive = true
	h.trainingPhase = "gpu-train"
	h.trainingStart = time.Now()
	h.trainingErr = ""
	h.trainingDone = false
	h.trainingResult = nil
	h.trainingMu.Unlock()

	go func() {
		err := h.runGPUTraining(svc)
		h.trainingMu.Lock()
		defer h.trainingMu.Unlock()
		h.trainingActive = false
		h.trainingDone = true
		if err != nil {
			h.trainingErr = err.Error()
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "GPU training started asynchronously (240 profiles)",
	})
}

// profileSampleJSON represents one profile sample for the Python training script.
type profileSampleJSON struct {
	Features    []float64 `json:"features"`
	FamilyLabel int       `json:"family_label"`
	BrowserType string    `json:"browser_type"`
}

type gpuTrainingPaths struct {
	InputPath    string
	OutputPath   string
	ProgressPath string
}

// exportProfileFeatures encodes all profiles to a JSON file for the Python GPU trainer.
func exportProfileFeatures(allProfiles []profiles.ClientProfile, outputPath string) (int, error) {
	labelMap := map[core.BrowserType]int{
		core.BrowserChrome:  0,
		core.BrowserFirefox: 1,
		core.BrowserSafari:  2,
		core.BrowserEdge:    3,
		core.BrowserOpera:   4,
		core.BrowserBrave:   5,
		core.BrowserSamsung: 6,
	}

	var samples []profileSampleJSON
	for i := range allProfiles {
		p := &allProfiles[i]
		features := ml.EncodeFingerprint(p)
		label, ok := labelMap[p.BrowserType]
		if !ok {
			label = 0
		}
		samples = append(samples, profileSampleJSON{
			Features:    features,
			FamilyLabel: label,
			BrowserType: string(p.BrowserType),
		})
	}

	data, err := json.Marshal(map[string]interface{}{"samples": samples})
	if err != nil {
		return 0, fmt.Errorf("marshal profile features: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
		return 0, fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return 0, fmt.Errorf("write profile features: %w", err)
	}
	return len(samples), nil
}

// runGPUTraining executes the Python GPU training script as a subprocess.
func (h *Handler) runGPUTraining(svc *ml.MLService) error {
	allProfiles := profiles.GetAll()
	paths := gpuTrainingPaths{
		InputPath:    "/tmp/ml_train_input.json",
		OutputPath:   "/models/weights.json",
		ProgressPath: "/tmp/ml_training_progress.json",
	}

	h.setTrainingPhase("exporting profiles")
	nProfiles, err := exportProfileFeatures(allProfiles, paths.InputPath)
	if err != nil {
		return fmt.Errorf("export profiles: %w", err)
	}

	defer os.Remove(paths.InputPath)
	defer os.Remove(paths.ProgressPath)

	h.setTrainingPhase(fmt.Sprintf("gpu-training (%d profiles)", nProfiles))
	trainOutput, err := runExternalGPUTraining(paths)
	if err != nil {
		return err
	}

	h.setTrainingPhase("loading weights")
	if err := svc.Pipeline().LoadWeights(paths.OutputPath); err != nil {
		return fmt.Errorf("load weights: %w", err)
	}

	result := h.buildGPUTrainingResult(svc, nProfiles, trainOutput, paths.ProgressPath)
	h.setTrainingResult(result)

	return nil
}

func (h *Handler) setTrainingPhase(phase string) {
	h.trainingMu.Lock()
	defer h.trainingMu.Unlock()
	h.trainingPhase = phase
}

func runExternalGPUTraining(paths gpuTrainingPaths) (string, error) {
	scriptPath := "/app/gpu_train.py"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		scriptPath = "training/gpu_train.py"
	}

	cmd := exec.Command("python3", scriptPath,
		"--input", paths.InputPath,
		"--output", paths.OutputPath,
		"--progress", paths.ProgressPath,
		"--epochs", "200",
	)
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gpu training failed: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

func (h *Handler) buildGPUTrainingResult(svc *ml.MLService, nProfiles int, output string, progressPath string) map[string]interface{} {
	st := svc.Stats()
	result := map[string]interface{}{
		"phase":         "gpu-train",
		"success":       true,
		"modelReady":    st.ModelReady,
		"modelVersions": st.ModelVersions,
		"profiles":      nProfiles,
		"output":        output,
	}
	if finalProgress := readGPUProgress(progressPath); finalProgress != nil {
		result["gpuProgress"] = finalProgress
	}
	return result
}

func (h *Handler) setTrainingResult(result map[string]interface{}) {
	h.trainingMu.Lock()
	defer h.trainingMu.Unlock()
	h.trainingResult = result
}

func readGPUProgress(progressPath string) map[string]interface{} {
	progressData, err := os.ReadFile(progressPath)
	if err != nil {
		return nil
	}

	var finalProgress map[string]interface{}
	if json.Unmarshal(progressData, &finalProgress) != nil {
		return nil
	}
	return finalProgress
}

// handleMLServiceTrainingStatus returns async training/evolution status.
func (h *Handler) handleMLServiceTrainingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.trainingMu.Lock()
	active := h.trainingActive
	phase := h.trainingPhase
	start := h.trainingStart
	done := h.trainingDone
	errMsg := h.trainingErr
	result := h.trainingResult
	h.trainingMu.Unlock()

	resp := map[string]interface{}{
		"active": active,
		"phase":  phase,
		"done":   done,
	}
	if !start.IsZero() {
		resp["elapsed"] = time.Since(start).Seconds()
		resp["startTime"] = start.Format(time.RFC3339)
	}
	if errMsg != "" {
		resp["error"] = errMsg
	}
	if result != nil {
		resp["result"] = result
	}

	// Read GPU training progress file if available
	if active {
		if progressData, err := os.ReadFile("/tmp/ml_training_progress.json"); err == nil {
			var gpuProgress map[string]interface{}
			if json.Unmarshal(progressData, &gpuProgress) == nil {
				resp["gpuProgress"] = gpuProgress
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleMLServiceFeedback submits feedback samples to online learning.
func (h *Handler) handleMLServiceFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		http.Error(w, "MLService not enabled", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		ProfileID string  `json:"profileId"`
		Label     string  `json:"label"`
		Reward    float64 `json:"reward"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var prof *profiles.ClientProfile
	if req.ProfileID != "" {
		if p, ok := h.findProfile(req.ProfileID); ok {
			prof = &p
		}
	}

	sample := &ml.FeedbackSample{
		Profile: prof,
		Label:   req.Label,
		Reward:  req.Reward,
	}
	svc.Feedback(sample)

	st := svc.Stats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"feedbackCount": st.FeedbackCount,
	})
}

// Ensure unused imports are referenced at package level.
var (
	_ = (*defense.RiskEngine)(nil)
	_ = (*frontend.SDK)(nil)
	_ = plugin.GetRegistryStats
	_ = (*tlsmod.JA3Result)(nil)
	_ = (*ml.MLService)(nil)
	_ = fmt.Sprintf
	_ = os.Remove
	_ = (*exec.Cmd)(nil)
	_ = filepath.Dir
)
