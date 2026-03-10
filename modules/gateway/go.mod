module github.com/vistone/fingerprint/modules/gateway

go 1.25.7

require (
	github.com/chromedp/cdproto v0.0.0-20241022234722-4d5d5faf59fb
	github.com/chromedp/chromedp v0.11.2
	github.com/vistone/fingerprint/modules/agent v1.0.8
	github.com/vistone/fingerprint/modules/core v1.0.8
	github.com/vistone/fingerprint/modules/defense v1.0.8
	github.com/vistone/fingerprint/modules/frontend v1.0.8
	github.com/vistone/fingerprint/modules/internal v1.0.8
	github.com/vistone/fingerprint/modules/ml v1.0.8
	github.com/vistone/fingerprint/modules/profiles v1.0.8
	google.golang.org/grpc v1.79.1
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/bogdanfinn/utls v1.7.7-barnius // indirect
	github.com/klauspost/compress v1.18.0 // indirect
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
