module github.com/vistone/fingerprint/modules/ml

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.4
	github.com/vistone/fingerprint/modules/profiles v1.0.4
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/profiles => ../profiles
)
