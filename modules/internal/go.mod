module github.com/vistone/fingerprint/modules/internal

go 1.25.7

require (
	github.com/bogdanfinn/utls v1.7.7-barnius
	github.com/prometheus/client_golang v1.23.2
	github.com/vistone/fingerprint/modules/core v1.0.6
	github.com/vistone/fingerprint/modules/defense v1.0.6
	github.com/vistone/fingerprint/modules/generator v1.0.6
	github.com/vistone/fingerprint/modules/http v1.0.6
	github.com/vistone/fingerprint/modules/ml v1.0.6
	github.com/vistone/fingerprint/modules/profiles v1.0.6
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
	go.uber.org/zap v1.27.1
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bogdanfinn/fhttp v0.6.8 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/vistone/fingerprint/modules/errors v1.0.6 // indirect
	github.com/vistone/fingerprint/modules/kit v1.0.6 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/ml => ../ml
	github.com/vistone/fingerprint/modules/profiles => ../profiles
	github.com/vistone/fingerprint/modules/tls => ../tls
)
