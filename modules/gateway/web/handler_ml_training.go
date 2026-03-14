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

	// Paths
	inputPath := "/tmp/ml_train_input.json"
	outputPath := "/models/weights.json"
	progressPath := "/tmp/ml_training_progress.json"

	// Phase 1: Export profile features
	h.trainingMu.Lock()
	h.trainingPhase = "exporting profiles"
	h.trainingMu.Unlock()

	nProfiles, err := exportProfileFeatures(allProfiles, inputPath)
	if err != nil {
		return fmt.Errorf("export profiles: %w", err)
	}

	// Clean up input file when done
	defer os.Remove(inputPath)
	defer os.Remove(progressPath)

	// Phase 2: Run Python GPU training
	h.trainingMu.Lock()
	h.trainingPhase = fmt.Sprintf("gpu-training (%d profiles)", nProfiles)
	h.trainingMu.Unlock()

	// Determine training script path
	scriptPath := "/app/gpu_train.py"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// Fallback for local development
		scriptPath = "training/gpu_train.py"
	}

	cmd := exec.Command("python3", scriptPath,
		"--input", inputPath,
		"--output", outputPath,
		"--progress", progressPath,
		"--epochs", "200",
	)
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	// Capture stdout/stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gpu training failed: %w\nOutput: %s", err, string(output))
	}

	// Phase 3: Load weights into pipeline
	h.trainingMu.Lock()
	h.trainingPhase = "loading weights"
	h.trainingMu.Unlock()

	if err := svc.Pipeline().LoadWeights(outputPath); err != nil {
		return fmt.Errorf("load weights: %w", err)
	}

	// Read final progress for result
	var finalProgress map[string]interface{}
	if progressData, err := os.ReadFile(progressPath); err == nil {
		json.Unmarshal(progressData, &finalProgress)
	}

	st := svc.Stats()
	result := map[string]interface{}{
		"phase":         "gpu-train",
		"success":       true,
		"modelReady":    st.ModelReady,
		"modelVersions": st.ModelVersions,
		"profiles":      nProfiles,
		"output":        string(output),
	}
	if finalProgress != nil {
		result["gpuProgress"] = finalProgress
	}

	h.trainingMu.Lock()
	h.trainingResult = result
	h.trainingMu.Unlock()

	return nil
}

// handleMLServiceTrainingStatus 查询异步训练/进化状态
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

// handleMLServiceFeedback 提交反馈样本给在线学习系统
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
