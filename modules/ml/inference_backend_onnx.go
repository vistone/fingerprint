package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

type onnxInferenceBackend struct {
	pipeline    *ModelPipeline
	pythonBin   string
	scriptPath  string
	manifest    string
	callTimeout time.Duration
}

type onnxInferRequest struct {
	Samples []onnxInferSample `json:"samples"`
}

type onnxInferSample struct {
	Features      []float64 `json:"features"`
	CrossFeatures []float64 `json:"cross_features"`
	Behavior      []float64 `json:"behavior"`
}

type onnxInferResponse struct {
	Results []onnxInferResult `json:"results"`
}

type onnxInferResult struct {
	Embedding    []float64 `json:"embedding"`
	FamilyLogits []float64 `json:"family_logits"`
	ForgeryProb  float64   `json:"forgery_prob"`
	TypeLogits   []float64 `json:"type_logits"`
	ThreatLogits []float64 `json:"threat_logits"`
	ActionLogits []float64 `json:"action_logits"`
}

func newONNXInferenceBackend(config *ServiceConfig, pipeline *ModelPipeline) *onnxInferenceBackend {
	if config == nil || config.ONNXModelDir == "" {
		return nil
	}

	manifestPath := config.ONNXModelDir + "/manifest.json"
	if _, err := os.Stat(manifestPath); err != nil {
		return nil
	}

	pythonBin := config.ONNXPythonBin
	if pythonBin == "" {
		pythonBin = "python3"
	}

	scriptPath := config.ONNXPythonScript
	if scriptPath == "" {
		scriptPath = "training/onnx_infer.py"
	}

	timeout := config.ONNXTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &onnxInferenceBackend{
		pipeline:    pipeline,
		pythonBin:   pythonBin,
		scriptPath:  scriptPath,
		manifest:    manifestPath,
		callTimeout: timeout,
	}
}

func (b *onnxInferenceBackend) Name() string {
	return "onnx"
}

func (b *onnxInferenceBackend) Infer(profile *profiles.ClientProfile, behavior []float64) (*PipelineResult, error) {
	features := EncodeFingerprint(profile)
	crossFeatures := ComputeCrossLayerFeatures(features)
	result, err := b.inferByFeatures(features, crossFeatures, behavior)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b *onnxInferenceBackend) InferFromFeatures(fv *core.FeatureVector, behavior []float64) (*PipelineResult, error) {
	features := EncodeFingerprintFromFeatureVector(fv)
	crossFeatures := ComputeCrossLayerFeatures(features)
	result, err := b.inferByFeatures(features, crossFeatures, behavior)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b *onnxInferenceBackend) InferBatch(profs []*profiles.ClientProfile, behaviors [][]float64) ([]*PipelineResult, error) {
	samples := make([]onnxInferSample, 0, len(profs))
	featureBatch := make([][]float64, 0, len(profs))
	crossBatch := make([][]float64, 0, len(profs))

	for index := range profs {
		features := EncodeFingerprint(profs[index])
		crossFeatures := ComputeCrossLayerFeatures(features)
		featureBatch = append(featureBatch, features)
		crossBatch = append(crossBatch, crossFeatures)
		samples = append(samples, onnxInferSample{
			Features:      features,
			CrossFeatures: crossFeatures,
			Behavior:      normalizeBehaviorFeatures(behaviors, index),
		})
	}

	response, err := b.runONNXInference(samples)
	if err != nil {
		return nil, err
	}
	if len(response.Results) != len(samples) {
		return nil, fmt.Errorf("onnx result count mismatch: got %d want %d", len(response.Results), len(samples))
	}

	results := make([]*PipelineResult, 0, len(samples))
	for index := range response.Results {
		result, buildErr := buildResultFromONNXOutputs(onnxOutputBundle{
			Features:      featureBatch[index],
			CrossFeatures: crossBatch[index],
			Embedding:     response.Results[index].Embedding,
			FamilyLogits:  response.Results[index].FamilyLogits,
			ForgeryProb:   response.Results[index].ForgeryProb,
			TypeLogits:    response.Results[index].TypeLogits,
			ThreatLogits:  response.Results[index].ThreatLogits,
			ActionLogits:  response.Results[index].ActionLogits,
		})
		if buildErr != nil {
			return nil, buildErr
		}
		results = append(results, result)
	}

	return results, nil
}

func (b *onnxInferenceBackend) inferByFeatures(features []float64, crossFeatures []float64, behavior []float64) (*PipelineResult, error) {
	response, err := b.runONNXInference([]onnxInferSample{{
		Features:      features,
		CrossFeatures: crossFeatures,
		Behavior:      normalizeBehavior(behavior),
	}})
	if err != nil {
		return nil, err
	}
	if len(response.Results) != 1 {
		return nil, fmt.Errorf("onnx infer expected 1 result, got %d", len(response.Results))
	}

	return buildResultFromONNXOutputs(onnxOutputBundle{
		Features:      features,
		CrossFeatures: crossFeatures,
		Embedding:     response.Results[0].Embedding,
		FamilyLogits:  response.Results[0].FamilyLogits,
		ForgeryProb:   response.Results[0].ForgeryProb,
		TypeLogits:    response.Results[0].TypeLogits,
		ThreatLogits:  response.Results[0].ThreatLogits,
		ActionLogits:  response.Results[0].ActionLogits,
	})
}

func (b *onnxInferenceBackend) runONNXInference(samples []onnxInferSample) (*onnxInferResponse, error) {
	request := onnxInferRequest{Samples: samples}
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal onnx request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.callTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, b.pythonBin, b.scriptPath, "--manifest", b.manifest)
	cmd.Stdin = bytes.NewReader(payload)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("onnx infer timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("onnx infer failed: %w; output: %s", err, string(output))
	}

	var response onnxInferResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("unmarshal onnx response: %w", err)
	}
	return &response, nil
}

func normalizeBehaviorFeatures(behaviors [][]float64, index int) []float64 {
	if index < len(behaviors) {
		return normalizeBehavior(behaviors[index])
	}
	return make([]float64, BehaviorFeatureDim)
}

func normalizeBehavior(behavior []float64) []float64 {
	out := make([]float64, BehaviorFeatureDim)
	if len(behavior) == 0 {
		return out
	}

	copy(out, behavior)
	return out
}
