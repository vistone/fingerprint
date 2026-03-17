package gateway

import (
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
	"github.com/vistone/fingerprint/modules/waf"
)

func TestClosedLoopEvaluateWithWAF_ActionAndScore(t *testing.T) {
	controller := NewClosedLoopController(&ClosedLoopConfig{Enabled: true}, nil)
	defer controller.Stop()

	profile := profiles.GetRandom()
	gen := &ml.GenerateResult{
		Profile:         &profile,
		SourceProfileID: profile.ID,
	}

	allowCfg := *waf.DefaultWAFConfig
	allowCfg.Enabled = false
	allowWAF := waf.NewWAF(&allowCfg)
	defer allowWAF.Stop()

	allowScore, allowAction := controller.evaluateWithWAF(gen, allowWAF)
	if allowAction != waf.ActionAllow {
		t.Fatalf("expected allow action, got %s", allowAction)
	}

	blockCfg := *waf.DefaultWAFConfig
	blockCfg.Enabled = true
	blockCfg.BlacklistIPs = []string{"198.51.100.10"}
	blockWAF := waf.NewWAF(&blockCfg)
	defer blockWAF.Stop()

	blockScore, blockAction := controller.evaluateWithWAF(gen, blockWAF)
	if blockAction != waf.ActionBlock {
		t.Fatalf("expected block action, got %s", blockAction)
	}
	if allowScore <= blockScore {
		t.Fatalf("expected allow score > block score, got allow=%f block=%f", allowScore, blockScore)
	}
}

func TestClosedLoopValidateAgainstWAF_RecordsFeedbackAndStats(t *testing.T) {
	svc, err := ml.NewMLService(&ml.ServiceConfig{
		ModelStorePath:             "",
		AutoLoadLatest:             false,
		FeedbackBufferSize:         32,
		ValidationForgeryThreshold: 1.0,
		ValidationConsistencyMin:   0.0,
	})
	if err != nil {
		t.Fatalf("create ML service failed: %v", err)
	}

	controller := NewClosedLoopController(&ClosedLoopConfig{
		Enabled:            true,
		TrainingInterval:   time.Hour,
		CandidatesPerCycle: 1,
		NoiseIntensity:     0.1,
	}, svc)
	defer controller.Stop()

	profile := profiles.GetRandom()
	candidates := []*ml.GenerateResult{
		{
			Profile:         &profile,
			SourceProfileID: profile.ID,
		},
	}

	wafCfg := *waf.DefaultWAFConfig
	wafCfg.BlacklistIPs = []string{"198.51.100.10"}
	wafInstance := waf.NewWAF(&wafCfg)
	defer wafInstance.Stop()

	before := svc.Stats().FeedbackCount
	controller.validateAgainstWAF(candidates, wafInstance, svc)
	after := svc.Stats().FeedbackCount

	if after != before+1 {
		t.Fatalf("expected feedback count +1, before=%d after=%d", before, after)
	}

	stats := controller.Stats()
	if stats.DetectionsProcessed != 1 {
		t.Fatalf("expected detections_processed=1, got %d", stats.DetectionsProcessed)
	}
}
