module github.com/vistone/fingerprint/examples/plugin

go 1.25.7

require github.com/vistone/fingerprint/modules/plugin v1.0.3

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/vistone/fingerprint/modules/internal v1.0.3 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
)

replace github.com/vistone/fingerprint/modules/plugin => ../../modules/plugin
