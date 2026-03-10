module github.com/vistone/fingerprint/modules/agent

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.4
	github.com/vistone/fingerprint/modules/defense v1.0.4
	github.com/vistone/fingerprint/modules/ml v1.0.4
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/defense => ../defense
	github.com/vistone/fingerprint/modules/ml => ../ml
	github.com/vistone/fingerprint/modules/http => ../http
	github.com/vistone/fingerprint/modules/internal => ../internal
	github.com/vistone/fingerprint/modules/network => ../network
	github.com/vistone/fingerprint/modules/profiles => ../profiles
	github.com/vistone/fingerprint/modules/tls => ../tls
)
