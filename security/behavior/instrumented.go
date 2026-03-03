package behavior

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// InstrumentedBehaviorAnalyzer 为 BehaviorAnalyzer 添加可观测能力
type InstrumentedBehaviorAnalyzer struct {
	inner   *BehaviorAnalyzer
	tracer  trace.Tracer
	logger  *zap.SugaredLogger
	metrics *BehaviorAnalysisMetrics
}

// BehaviorAnalysisMetrics 行为分析指标
type BehaviorAnalysisMetrics struct {
	AddRequestCount       int64
	AnalysisCount         int64
	AnomalyDetectionCount int64
	SignalGenerationCount int64
	LastAnalysisDuration  time.Duration
	TotalAnalysisDuration time.Duration
}

// NewInstrumentedBehaviorAnalyzer 创建可观测的行为分析器
func NewInstrumentedBehaviorAnalyzer(
	analyzer *BehaviorAnalyzer,
	tracer trace.Tracer,
	logger *zap.SugaredLogger,
) *InstrumentedBehaviorAnalyzer {
	if tracer == nil {
		tracer = otel.Tracer("behavior-analysis")
	}

	return &InstrumentedBehaviorAnalyzer{
		inner:   analyzer,
		tracer:  tracer,
		logger:  logger,
		metrics: &BehaviorAnalysisMetrics{},
	}
}

// AddRequest 添加请求行为
func (iba *InstrumentedBehaviorAnalyzer) AddRequest(ctx context.Context, req RequestBehavior) {
	_, span := iba.tracer.Start(ctx, "BehaviorAnalyzer.AddRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("source_ip", req.SourceIP),
		attribute.String("sni", req.SNI),
		attribute.String("tls_version", req.TLSVersion),
		attribute.String("http_version", req.HTTPVersion),
		attribute.Bool("reusing_connection", req.ReusingConnection),
	)

	iba.metrics.AddRequestCount++
	iba.inner.AddRequest(req)

	if iba.logger != nil {
		iba.logger.Debugw("request recorded",
			"source_ip", req.SourceIP,
			"sni", req.SNI,
			"tls_version", req.TLSVersion,
		)
	}
}

// AnalyzeTemporalPattern 分析时序模式
func (iba *InstrumentedBehaviorAnalyzer) AnalyzeTemporalPattern(ctx context.Context, origin string) *TemporalPattern {
	ctx, span := iba.tracer.Start(ctx, "BehaviorAnalyzer.AnalyzeTemporalPattern")
	defer span.End()

	startTime := time.Now()
	span.SetAttributes(attribute.String("origin", origin))

	pattern := iba.inner.AnalyzeTemporalPattern(origin)

	duration := time.Since(startTime)
	iba.recordAnalysisDuration(duration)

	if pattern != nil {
		span.SetAttributes(
			attribute.Int("interval_count", len(pattern.Intervals)),
			attribute.Float64("regularity_index", pattern.RegularityIndex),
			attribute.Int("anomalous_intervals", pattern.AnomalousIntervals),
		)

		if iba.logger != nil {
			iba.logger.Infow("temporal pattern analyzed",
				"origin", origin,
				"interval_count", len(pattern.Intervals),
				"regularity_index", pattern.RegularityIndex,
				"duration_ms", duration.Milliseconds(),
			)
		}
	}

	return pattern
}

// AnalyzeProtocolProportion 分析协议比例
func (iba *InstrumentedBehaviorAnalyzer) AnalyzeProtocolProportion(ctx context.Context, origin string) *ProtocolProportion {
	ctx, span := iba.tracer.Start(ctx, "BehaviorAnalyzer.AnalyzeProtocolProportion")
	defer span.End()

	startTime := time.Now()
	span.SetAttributes(attribute.String("origin", origin))

	prop := iba.inner.AnalyzeProtocolProportion(origin)

	duration := time.Since(startTime)
	iba.recordAnalysisDuration(duration)

	if prop != nil {
		span.SetAttributes(
			attribute.Float64("entropy_score", prop.EntropyScore),
			attribute.Bool("is_anomalous", prop.IsAnomalous),
		)

		if iba.logger != nil {
			iba.logger.Infow("protocol proportion analyzed",
				"origin", origin,
				"entropy_score", prop.EntropyScore,
				"is_anomalous", prop.IsAnomalous,
				"duration_ms", duration.Milliseconds(),
			)
		}
	}

	return prop
}

// GenerateBehaviorSignals 生成行为信号
func (iba *InstrumentedBehaviorAnalyzer) GenerateBehaviorSignals(ctx context.Context, origin string) []BehaviorSignal {
	ctx, span := iba.tracer.Start(ctx, "BehaviorAnalyzer.GenerateBehaviorSignals")
	defer span.End()

	startTime := time.Now()
	span.SetAttributes(attribute.String("origin", origin))

	signals := iba.inner.GenerateBehaviorSignals(origin)

	duration := time.Since(startTime)
	iba.recordAnalysisDuration(duration)
	if len(signals) > 0 {
		iba.metrics.AnomalyDetectionCount++
	}

	// 统计高风险信号
	var criticalCount, highCount int
	for _, sig := range signals {
		if sig.RiskLevel == "critical" {
			criticalCount++
		} else if sig.RiskLevel == "high" {
			highCount++
		}
	}

	span.SetAttributes(
		attribute.Int("signal_count", len(signals)),
		attribute.Int("critical_signals", criticalCount),
		attribute.Int("high_signals", highCount),
	)

	if iba.logger != nil {
		iba.logger.Infow("behavior signals generated",
			"origin", origin,
			"signal_count", len(signals),
			"critical", criticalCount,
			"high", highCount,
			"duration_ms", duration.Milliseconds(),
		)
	}

	return signals
}

// GetAllSignals 获取所有信号
func (iba *InstrumentedBehaviorAnalyzer) GetAllSignals(ctx context.Context) []BehaviorSignal {
	_, span := iba.tracer.Start(ctx, "BehaviorAnalyzer.GetAllSignals")
	defer span.End()

	signals := iba.inner.GetAllSignals()
	span.SetAttributes(attribute.Int("signal_count", len(signals)))

	return signals
}

// GetRiskScore 获取风险评分
func (iba *InstrumentedBehaviorAnalyzer) GetRiskScore(ctx context.Context) float64 {
	_, span := iba.tracer.Start(ctx, "BehaviorAnalyzer.GetRiskScore")
	defer span.End()

	score := iba.inner.GetRiskScore()
	span.SetAttributes(attribute.Float64("risk_score", score))

	return score
}

// GetAnalysisSummary 获取分析摘要
func (iba *InstrumentedBehaviorAnalyzer) GetAnalysisSummary(ctx context.Context) string {
	_, span := iba.tracer.Start(ctx, "BehaviorAnalyzer.GetAnalysisSummary")
	defer span.End()

	summary := iba.inner.GetAnalysisSummary()
	span.AddEvent("analysis_summary_retrieved")

	if iba.logger != nil {
		iba.logger.Debugw("analysis summary", "summary", summary)
	}

	return summary
}

// PassthroughToInner 获取底层分析器以访问原始接口
func (iba *InstrumentedBehaviorAnalyzer) PassthroughToInner() *BehaviorAnalyzer {
	return iba.inner
}

// GetMetrics 获取指标快照
func (iba *InstrumentedBehaviorAnalyzer) GetMetrics() *BehaviorAnalysisMetrics {
	return iba.metrics
}

// SetLogger 设置日志记录器
func (iba *InstrumentedBehaviorAnalyzer) SetLogger(logger *zap.SugaredLogger) {
	iba.logger = logger
}

// SetTracer 设置追踪器
func (iba *InstrumentedBehaviorAnalyzer) SetTracer(tracer trace.Tracer) {
	iba.tracer = tracer
}

// recordAnalysisDuration 记录分析耗时
func (iba *InstrumentedBehaviorAnalyzer) recordAnalysisDuration(d time.Duration) {
	iba.metrics.AnalysisCount++
	iba.metrics.LastAnalysisDuration = d
	iba.metrics.TotalAnalysisDuration += d
}
