package gateway

import (
	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/plugin"
)

// Close gracefully shuts down the gateway and releases background resources.
func (g *Gateway) Close() {
	if g.limiter != nil {
		g.limiter.Close()
	}
	if g.crawler != nil {
		g.crawler.Stop()
	}
	if g.waf != nil {
		g.waf.Stop()
	}
	if g.agent != nil {
		g.agent.Stop()
	}
}

// GetAgent returns the autonomous security agent instance (for web admin console queries).
func (g *Gateway) GetAgent() *agent.Agent {
	return g.agent
}

// GetClassifier returns the ML hierarchical classifier.
func (g *Gateway) GetClassifier() *ml.HierarchicalClassifier {
	return g.classifier
}

// GetExtractor returns the feature extractor.
func (g *Gateway) GetExtractor() *ml.FeatureExtractor {
	return g.extractor
}

// GetRiskEngine returns the risk assessment engine.
func (g *Gateway) GetRiskEngine() *defense.RiskEngine {
	return g.riskEngine
}

// GetSDK returns the frontend SDK.
func (g *Gateway) GetSDK() *frontend.SDK {
	return g.sdk
}

// GetInjector returns the HTML injector.
func (g *Gateway) GetInjector() *HTMLInjector {
	return g.injector
}

// GetProfileManager returns the profile manager.
func (g *Gateway) GetProfileManager() *ProfileManager {
	return g.profileManager
}

// GetMLService returns the central ML service (nil if not enabled).
func (g *Gateway) GetMLService() *ml.MLService {
	return g.mlService
}

// GetPluginManager returns the plugin manager.
func (g *Gateway) GetPluginManager() *plugin.Manager {
	return g.pluginManager
}
