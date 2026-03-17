package ml

import (
	"fmt"
	"math"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

type inferenceBackend interface {
	Name() string
	Infer(profile *profiles.ClientProfile, behavior []float64) (*PipelineResult, error)
	InferFromFeatures(fv *core.FeatureVector, behavior []float64) (*PipelineResult, error)
	InferBatch(profs []*profiles.ClientProfile, behaviors [][]float64) ([]*PipelineResult, error)
}

type nativeInferenceBackend struct {
	pipeline *ModelPipeline
}

func (b *nativeInferenceBackend) Name() string {
	return "native"
}

func (b *nativeInferenceBackend) Infer(profile *profiles.ClientProfile, behavior []float64) (*PipelineResult, error) {
	return b.pipeline.Infer(profile, behavior), nil
}

func (b *nativeInferenceBackend) InferFromFeatures(fv *core.FeatureVector, behavior []float64) (*PipelineResult, error) {
	return b.pipeline.InferFromFeatures(fv, behavior), nil
}

func (b *nativeInferenceBackend) InferBatch(profs []*profiles.ClientProfile, behaviors [][]float64) ([]*PipelineResult, error) {
	return b.pipeline.InferBatch(profs, behaviors), nil
}

func newInferenceBackend(config *ServiceConfig, pipeline *ModelPipeline) inferenceBackend {
	if config != nil && config.InferenceBackend == "onnx" {
		backend := newONNXInferenceBackend(config, pipeline)
		if backend != nil {
			return backend
		}
	}

	return &nativeInferenceBackend{pipeline: pipeline}
}

type onnxOutputBundle struct {
	Features      []float64
	CrossFeatures []float64
	Embedding     []float64
	FamilyLogits  []float64
	ForgeryProb   float64
	TypeLogits    []float64
	ThreatLogits  []float64
	ActionLogits  []float64
}

func buildResultFromONNXOutputs(outputs onnxOutputBundle) (*PipelineResult, error) {
	if len(outputs.Embedding) != EmbeddingDim {
		return nil, fmt.Errorf("onnx embedding dim mismatch: got %d want %d", len(outputs.Embedding), EmbeddingDim)
	}
	if len(outputs.FamilyLogits) != NumBrowserFamilies {
		return nil, fmt.Errorf("onnx family logits mismatch: got %d want %d", len(outputs.FamilyLogits), NumBrowserFamilies)
	}
	if len(outputs.TypeLogits) != NumForgeryTypes {
		return nil, fmt.Errorf("onnx type logits mismatch: got %d want %d", len(outputs.TypeLogits), NumForgeryTypes)
	}
	if len(outputs.ThreatLogits) != NumThreatClasses {
		return nil, fmt.Errorf("onnx threat logits mismatch: got %d want %d", len(outputs.ThreatLogits), NumThreatClasses)
	}
	if len(outputs.ActionLogits) != NumActions {
		return nil, fmt.Errorf("onnx action logits mismatch: got %d want %d", len(outputs.ActionLogits), NumActions)
	}

	browserProbs := softmax(outputs.FamilyLogits)
	browserIdx := argmax(browserProbs)
	browserNames := make([]core.BrowserType, NumBrowserFamilies)
	copy(browserNames, indexFamily[:])
	browser := BrowserPrediction{
		Family:      indexFamily[browserIdx],
		Confidence:  browserProbs[browserIdx],
		Probs:       browserProbs,
		FamilyNames: browserNames,
	}

	typeProbs := softmax(outputs.TypeLogits)
	typeIdx := argmax(typeProbs)
	typeNames := make([]string, NumForgeryTypes)
	copy(typeNames, forgeryTypeNames[:])
	forgery := ForgeryResult{
		IsForgery:   outputs.ForgeryProb > 0.5,
		ForgeryProb: outputs.ForgeryProb,
		ForgeryType: ForgeryType(typeIdx),
		TypeProbs:   typeProbs,
		TypeNames:   typeNames,
	}

	threatProbs := softmax(outputs.ThreatLogits)
	actionProbs := softmax(outputs.ActionLogits)
	threatIdx := argmax(threatProbs)
	actionIdx := argmax(actionProbs)
	threat := ThreatPrediction{
		ThreatClass:      ThreatClass(threatIdx),
		ThreatProb:       1.0 - threatProbs[0],
		Action:           ActionClass(actionIdx),
		ActionConfidence: actionProbs[actionIdx],
		ClassProbs:       threatProbs,
		ActionProbs:      actionProbs,
	}

	return &PipelineResult{
		Embedding:     outputs.Embedding,
		Browser:       browser,
		Forgery:       forgery,
		Threat:        threat,
		RawFeatures:   outputs.Features,
		CrossFeatures: outputs.CrossFeatures,
	}, nil
}

func softmax(logits []float64) []float64 {
	if len(logits) == 0 {
		return nil
	}

	maxVal := logits[0]
	for i := 1; i < len(logits); i++ {
		if logits[i] > maxVal {
			maxVal = logits[i]
		}
	}

	expVals := make([]float64, len(logits))
	sumExp := 0.0
	for i := range logits {
		expVals[i] = math.Exp(logits[i] - maxVal)
		sumExp += expVals[i]
	}

	if sumExp <= 0 {
		fallback := make([]float64, len(logits))
		fallback[0] = 1.0
		return fallback
	}

	for i := range expVals {
		expVals[i] /= sumExp
	}
	return expVals
}

func argmax(values []float64) int {
	if len(values) == 0 {
		return 0
	}

	maxIdx := 0
	maxVal := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] > maxVal {
			maxVal = values[i]
			maxIdx = i
		}
	}
	return maxIdx
}
