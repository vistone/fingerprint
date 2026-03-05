module github.com/vistone/fingerprint/modules/http

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v0.0.0
	github.com/vistone/fingerprint/modules/profiles v0.0.0
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/bogdanfinn/fhttp v0.6.3 // indirect
	github.com/bogdanfinn/utls v1.7.4-barnius // indirect
	github.com/cloudflare/circl v1.5.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/profiles => ../profiles
)
