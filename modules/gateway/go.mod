module github.com/vistone/fingerprint/modules/gateway

go 1.25.7

require (
	github.com/chromedp/cdproto v0.0.0-20241022234722-4d5d5faf59fb
	github.com/chromedp/chromedp v0.11.2
	github.com/vistone/fingerprint v1.0.11
	github.com/vistone/fingerprint/modules/agent v1.0.11
	github.com/vistone/fingerprint/modules/core v1.0.11
	github.com/vistone/fingerprint/modules/defense v1.0.11
	github.com/vistone/fingerprint/modules/errors v1.0.11
	github.com/vistone/fingerprint/modules/frontend v1.0.11
	github.com/vistone/fingerprint/modules/internal v1.0.11
	github.com/vistone/fingerprint/modules/kit v1.0.11
	github.com/vistone/fingerprint/modules/ml v1.0.11
	github.com/vistone/fingerprint/modules/plugin v1.0.11
	github.com/vistone/fingerprint/modules/profiles v1.0.11
	github.com/vistone/fingerprint/modules/tls v1.0.11
	google.golang.org/grpc v1.79.1
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/bogdanfinn/fhttp v0.6.8 // indirect
	github.com/bogdanfinn/utls v1.7.7-barnius // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace (
	github.com/vistone/fingerprint/modules/agent => ../agent
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/defense => ../defense
	github.com/vistone/fingerprint/modules/frontend => ../frontend
	github.com/vistone/fingerprint/modules/ml => ../ml
	github.com/vistone/fingerprint/modules/profiles => ../profiles
)
