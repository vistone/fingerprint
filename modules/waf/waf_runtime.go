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

// RecentDecisions returns recent WAF decisions in reverse-chronological order.
func (w *WAF) RecentDecisions(limit int) []WAFDecision {
	if w == nil {
		return nil
	}
	if limit <= 0 {
		limit = 10
	}

	w.decisionMu.RLock()
	defer w.decisionMu.RUnlock()

	n := len(w.recentDecisions)
	if n == 0 {
		return nil
	}
	if limit > n {
		limit = n
	}

	out := make([]WAFDecision, limit)
	copy(out, w.recentDecisions[n-limit:])

	// Reverse to return newest first.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	return out
}

func (w *WAF) recordDecision(decision WAFDecision) {
	if w == nil {
		return
	}

	w.decisionMu.Lock()
	defer w.decisionMu.Unlock()

	w.recentDecisions = append(w.recentDecisions, decision)
	if len(w.recentDecisions) > 64 {
		w.recentDecisions = w.recentDecisions[len(w.recentDecisions)-64:]
	}
}
