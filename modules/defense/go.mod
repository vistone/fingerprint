module github.com/vistone/fingerprint/modules/defense

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.3
	github.com/vistone/fingerprint/modules/ml v1.0.3
	github.com/vistone/fingerprint/modules/profiles v1.0.3
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/bogdanfinn/fhttp v0.6.3 // indirect
	github.com/bogdanfinn/utls v1.7.4-barnius // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudflare/circl v1.5.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/ml => ../ml
	github.com/vistone/fingerprint/modules/profiles => ../profiles
)
