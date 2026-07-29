package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/ml"
)

func buildGatewayConfig(port int) gateway.GatewayConfig {
	config := *gateway.DefaultGatewayConfig
	config.Port = port
	applyGatewayEnv(&config)
	applyClosedLoopEnv(&config)

	return config
}

func applyGatewayEnv(config *gateway.GatewayConfig) {
	applyAntiDetectEnv(config)
	applyScannerEnv(config)
	applyMLEnv(config)
	applyRiskAndProxyEnv(config)
}

func applyAntiDetectEnv(config *gateway.GatewayConfig) {
	if parsed, ok := readEnvBool("FP_ANTIDETECT_ENABLED"); ok {
		config.AntiDetectEnabled = parsed
	}

	if profileID, ok := readEnvString("FP_ANTIDETECT_PROFILE"); ok {
		config.AntiDetectProfileID = profileID
	}

	if configDir, ok := readEnvString("FP_ANTIDETECT_CONFIG_DIR"); ok {
		config.AntiDetectConfigDir = configDir
	}

	if proxyTarget, ok := readEnvString("FP_ANTIDETECT_PROXY_TARGET"); ok {
		config.AntiDetectProxyTarget = proxyTarget
	}

	if parsed, ok := readEnvBool("FP_ANTIDETECT_INJECT_CONSISTENCY"); ok {
		config.AntiDetectInjectConsist = parsed
	}

	if parsed, ok := readEnvBool("FP_ANTIDETECT_DIRECT_PROXY"); ok {
		config.AntiDetectDirectProxy = parsed
	}
}

func applyScannerEnv(config *gateway.GatewayConfig) {
	if parsed, ok := readEnvBool("FP_SCANNER_USE_BROWSER"); ok {
		config.ScannerUseBrowser = parsed
	}

	if browserWS, ok := readEnvString("FP_SCANNER_BROWSER_WS"); ok {
		config.ScannerBrowserWS = browserWS
	}

	if timeoutMilliseconds, ok := readEnvInt("FP_SCANNER_BROWSER_TIMEOUT_MS"); ok && timeoutMilliseconds > 0 {
		config.ScannerBrowserTimeout = time.Duration(timeoutMilliseconds) * time.Millisecond
	}
}

func applyMLEnv(config *gateway.GatewayConfig) {
	if parsed, ok := readEnvBool("FP_ML_SERVICE_ENABLED"); ok {
		config.MLServiceEnabled = parsed
	}

	if classifierPath, ok := readEnvString("FP_ML_CLASSIFIER_PATH"); ok {
		config.MLClassifierPath = classifierPath
	}

	if trainingDataPath, ok := readEnvString("FP_ML_TRAINING_DATA"); ok {
		config.MLTrainingData = trainingDataPath
	}

	if backend, ok := readEnvString("FP_ML_INFERENCE_BACKEND"); ok {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.InferenceBackend = backend
	}

	if onnxDir, ok := readEnvString("FP_ML_ONNX_MODEL_DIR"); ok {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.ONNXModelDir = onnxDir
	}

	if pythonBin, ok := readEnvString("FP_ML_ONNX_PYTHON_BIN"); ok {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.ONNXPythonBin = pythonBin
	}

	if scriptPath, ok := readEnvString("FP_ML_ONNX_PYTHON_SCRIPT"); ok {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.ONNXPythonScript = scriptPath
	}

	if timeoutMilliseconds, ok := readEnvInt("FP_ML_ONNX_TIMEOUT_MS"); ok && timeoutMilliseconds > 0 {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.ONNXTimeout = time.Duration(timeoutMilliseconds) * time.Millisecond
	}

	if parsed, ok := readEnvBool("FP_ML_SHADOW_COMPARE_ENABLED"); ok {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.ShadowCompareEnabled = parsed
	}

	if rate, ok := readEnvFloat("FP_ML_SHADOW_SAMPLE_RATE"); ok && rate >= 0 && rate <= 1 {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.ShadowSampleRate = rate
	}

	if metricsPath, ok := readEnvString("FP_ML_SHADOW_METRICS_PATH"); ok {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.ShadowMetricsPath = metricsPath
	}

	if parsed, ok := readEnvBool("FP_ML_CANARY_ENABLED"); ok {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.CanaryEnabled = parsed
	}

	if rate, ok := readEnvFloat("FP_ML_CANARY_RATE"); ok && rate >= 0 && rate <= 1 {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.CanaryRate = rate
	}

	if backend, ok := readEnvString("FP_ML_CANARY_BACKEND"); ok {
		ensureMLServiceConfig(config)
		config.MLServiceConfig.CanaryBackend = backend
	}
}

func applyRiskAndProxyEnv(config *gateway.GatewayConfig) {
	if threshold, ok := readEnvFloat("FP_RISK_THRESHOLD"); ok && threshold > 0 {
		config.RiskThreshold = threshold
	}

	if apiKeys, ok := readEnvString("FP_API_KEYS"); ok {
		config.APIKeys = splitCSV(apiKeys)
	}

	if trustedProxies, ok := readEnvString("FP_TRUSTED_PROXIES"); ok {
		config.TrustedProxies = splitCSV(trustedProxies)
	}

	if pluginConfigPath, ok := readEnvString("FP_PLUGIN_CONFIG_PATH"); ok {
		config.PluginConfigPath = pluginConfigPath
	}
}

func applyClosedLoopEnv(config *gateway.GatewayConfig) {
	if parsed, ok := readEnvBool("FP_CLOSED_LOOP_ENABLED"); ok {
		config.ClosedLoopEnabled = parsed
	}

	if !config.ClosedLoopEnabled {
		return
	}

	closedLoopConfig := &gateway.ClosedLoopConfig{
		Enabled:            true,
		TrainingInterval:   time.Hour,
		CandidatesPerCycle: closedLoopDefaultCandidates,
		NoiseIntensity:     closedLoopDefaultNoise,
	}

	if duration, ok := readEnvDuration("FP_CLOSED_LOOP_TRAINING_INTERVAL"); ok && duration > 0 {
		closedLoopConfig.TrainingInterval = duration
	}

	if candidates, ok := readEnvInt("FP_CLOSED_LOOP_CANDIDATES"); ok && candidates > 0 {
		closedLoopConfig.CandidatesPerCycle = candidates
	}

	if noise, ok := readEnvFloat("FP_CLOSED_LOOP_NOISE"); ok && noise >= 0 && noise <= 1 {
		closedLoopConfig.NoiseIntensity = noise
	}

	config.ClosedLoopConfig = closedLoopConfig
}

func readEnvString(name string) (string, bool) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", false
	}

	return value, true
}

func readEnvBool(name string) (bool, bool) {
	value, ok := readEnvString(name)
	if !ok {
		return false, false
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}

	return parsed, true
}

func readEnvInt(name string) (int, bool) {
	value, ok := readEnvString(name)
	if !ok {
		return 0, false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func readEnvFloat(name string) (float64, bool) {
	value, ok := readEnvString(name)
	if !ok {
		return 0, false
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func readEnvDuration(name string) (time.Duration, bool) {
	value, ok := readEnvString(name)
	if !ok {
		return 0, false
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func ensureMLServiceConfig(config *gateway.GatewayConfig) {
	if config.MLServiceConfig != nil {
		return
	}

	defaultClone := *ml.DefaultServiceConfig
	config.MLServiceConfig = &defaultClone
}
