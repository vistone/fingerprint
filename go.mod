module github.com/vistone/fingerprint

go 1.25.4

require (
	github.com/bogdanfinn/fhttp v0.6.3
	github.com/bogdanfinn/utls v1.7.4-barnius
)

replace (
	github.com/vistone/domaindns => ../domaindns
	github.com/vistone/localippool => ../localippool
	github.com/vistone/logs => ../logs
	github.com/vistone/netconnpool => ../netconnpool
	github.com/vistone/quic => ../quic
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
)
