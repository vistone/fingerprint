package ml

import (
	"encoding/json"
	"fmt"
	"os"
)

// =========================================================================
// Model serialization — save/load trained model weights
// =========================================================================

// ModelWeights holds serialized model weights.
type ModelWeights struct {
	Version     string            `json:"version"`
	Encoder     []SerializedParam `json:"encoder"`
	Classifier  []SerializedParam `json:"classifier"`
	DetectorNet []SerializedParam `json:"detector_net"`
	TypeNet     []SerializedParam `json:"type_net"`
	ThreatNet   []SerializedParam `json:"threat_net"`
	ActionNet   []SerializedParam `json:"action_net"`
	Metrics     []TrainingMetrics `json:"metrics,omitempty"`
}

// SerializedParam holds a serialized parameter.
type SerializedParam struct {
	Shape []int     `json:"shape"`
	Data  []float64 `json:"data"`
}

// SaveWeights saves model weights to a file.
func (p *ModelPipeline) SaveWeights(path string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	weights := &ModelWeights{
		Version:     "1.0.14",
		Encoder:     serializeParams(p.encoder.Net.Params()),
		Classifier:  serializeParams(p.classifier.Net.Params()),
		DetectorNet: serializeParams(p.detector.DetectorNet.Params()),
		TypeNet:     serializeParams(p.detector.TypeNet.Params()),
		ThreatNet:   serializeParams(p.assessor.ThreatNet.Params()),
		ActionNet:   serializeParams(p.assessor.ActionNet.Params()),
	}

	data, err := json.Marshal(weights)
	if err != nil {
		return fmt.Errorf("marshal weights: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// LoadWeights loads model weights from a file.
func (p *ModelPipeline) LoadWeights(path string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read weights: %w", err)
	}

	var weights ModelWeights
	if err := json.Unmarshal(data, &weights); err != nil {
		return fmt.Errorf("unmarshal weights: %w", err)
	}

	deserializeParams(p.encoder.Net.Params(), weights.Encoder)
	deserializeParams(p.classifier.Net.Params(), weights.Classifier)
	deserializeParams(p.detector.DetectorNet.Params(), weights.DetectorNet)
	deserializeParams(p.detector.TypeNet.Params(), weights.TypeNet)
	deserializeParams(p.assessor.ThreatNet.Params(), weights.ThreatNet)
	deserializeParams(p.assessor.ActionNet.Params(), weights.ActionNet)

	p.trained = true
	return nil
}

func serializeParams(params []*Param) []SerializedParam {
	out := make([]SerializedParam, len(params))
	for i, p := range params {
		out[i] = SerializedParam{
			Shape: append([]int{}, p.Value.Shape...),
			Data:  append([]float64{}, p.Value.Data...),
		}
	}
	return out
}

func deserializeParams(params []*Param, serialized []SerializedParam) {
	for i := range params {
		if i >= len(serialized) {
			break
		}
		s := serialized[i]
		if len(s.Data) == len(params[i].Value.Data) {
			copy(params[i].Value.Data, s.Data)
			params[i].Value.Shape = append([]int{}, s.Shape...)
		}
	}
}

// =========================================================================
// Helper functions
// =========================================================================
