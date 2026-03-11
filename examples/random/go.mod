module github.com/vistone/fingerprint/examples/random

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.11
	github.com/vistone/fingerprint/modules/profiles v1.0.11
)

replace (
	github.com/vistone/fingerprint/modules/generator => ../../modules/generator
	github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles
)
