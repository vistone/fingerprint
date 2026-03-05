module github.com/vistone/fingerprint/examples/advanced

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v0.0.0
	github.com/vistone/fingerprint/modules/fingerprint v0.0.0
	github.com/vistone/fingerprint/modules/profiles v0.0.0
)

replace (
	github.com/vistone/fingerprint/modules/core => ../../modules/core
	github.com/vistone/fingerprint/modules/fingerprint => ../../modules/fingerprint
	github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles
)
