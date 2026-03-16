package waf

import "github.com/vistone/fingerprint/modules/ml"

// SetMLService attaches the shared gateway ML service to the WAF.
func (w *WAF) SetMLService(svc *ml.MLService) {
	if w == nil || svc == nil {
		return
	}
	w.mlService = svc
	w.learningPipeline = NewLearningPipeline(svc)
}

// ConfigSnapshot returns a copy of the WAF configuration.
func (w *WAF) ConfigSnapshot() WAFConfig {
	if w == nil || w.config == nil {
		return WAFConfig{}
	}
	return *w.config
}
