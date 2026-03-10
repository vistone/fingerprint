module github.com/vistone/fingerprint/modules/http

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.6
	github.com/vistone/fingerprint/modules/errors v1.0.6
	github.com/vistone/fingerprint/modules/internal v1.0.6
	github.com/vistone/fingerprint/modules/kit v1.0.6
	github.com/vistone/fingerprint/modules/profiles v1.0.6
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bogdanfinn/fhttp v0.6.8 // indirect
	github.com/bogdanfinn/utls v1.7.7-barnius // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/profiles => ../profiles
)
