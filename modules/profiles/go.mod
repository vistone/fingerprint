module github.com/vistone/fingerprint/modules/profiles

go 1.25.7

require (
	github.com/bogdanfinn/fhttp v0.6.8
	github.com/bogdanfinn/utls v1.7.7-barnius
	github.com/vistone/fingerprint/modules/core v1.0.6
	github.com/vistone/fingerprint/modules/errors v1.0.6
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
)

replace github.com/vistone/fingerprint/modules/core => ../core
