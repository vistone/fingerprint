module github.com/vistone/fingerprint/modules/tls

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v0.0.0
	github.com/vistone/fingerprint/modules/profiles v0.0.0
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/profiles => ../profiles
)
