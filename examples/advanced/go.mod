module github.com/vistone/fingerprint/examples/advanced

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.3
	github.com/vistone/fingerprint/modules/fingerprint v1.0.3
	github.com/vistone/fingerprint/modules/profiles v1.0.3
)

replace (
	github.com/vistone/fingerprint/modules/core => ../../modules/core
	github.com/vistone/fingerprint/modules/fingerprint => ../../modules/fingerprint
	github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles
)
