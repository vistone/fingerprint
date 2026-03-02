package defense

import fp "github.com/vistone/fingerprint"

// AnomalyDetector 异常检测器。
type AnomalyDetector = fp.AnomalyDetector

// ContradictionDetector 矛盾检测器。
type ContradictionDetector = fp.ContradictionDetector

// PassiveRecognizer 被动识别器。
type PassiveRecognizer = fp.PassiveRecognizer

// RecognitionResult 识别结果。
type RecognitionResult = fp.RecognitionResult

// NewAnomalyDetector 创建异常检测器。
func NewAnomalyDetector() *AnomalyDetector {
	return fp.NewAnomalyDetector()
}

// NewContradictionDetector 创建矛盾检测器。
func NewContradictionDetector() *ContradictionDetector {
	return fp.NewContradictionDetector()
}

// NewPassiveRecognizer 创建被动识别器。
func NewPassiveRecognizer() *PassiveRecognizer {
	return fp.NewPassiveRecognizer()
}

// RecognizeFromUserAgent 根据 User-Agent 便捷识别。
func RecognizeFromUserAgent(userAgent string) *RecognitionResult {
	return fp.RecognizeFromUserAgent(userAgent)
}
