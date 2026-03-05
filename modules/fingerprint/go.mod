module github.com/vistone/fingerprint/modules/fingerprint

go 1.25.7

require (
	github.com/bogdanfinn/utls v1.7.4-barnius
	github.com/vistone/fingerprint/modules/core v1.0.3
	github.com/vistone/fingerprint/modules/defense v1.0.3
	github.com/vistone/fingerprint/modules/frontend v1.0.3
	github.com/vistone/fingerprint/modules/gateway v1.0.3
	github.com/vistone/fingerprint/modules/http v1.0.3
	github.com/vistone/fingerprint/modules/ml v1.0.3
	github.com/vistone/fingerprint/modules/profiles v1.0.3
	github.com/vistone/fingerprint/modules/tls v1.0.3
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/bogdanfinn/fhttp v0.6.3 // indirect
	github.com/cloudflare/circl v1.5.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.79.1 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/defense => ../defense
	github.com/vistone/fingerprint/modules/frontend => ../frontend
	github.com/vistone/fingerprint/modules/gateway => ../gateway
	github.com/vistone/fingerprint/modules/http => ../http
	github.com/vistone/fingerprint/modules/ml => ../ml
	github.com/vistone/fingerprint/modules/profiles => ../profiles
	github.com/vistone/fingerprint/modules/tls => ../tls
)
