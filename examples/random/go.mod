module github.com/vistone/fingerprint/examples/random

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/generator v0.0.0
	github.com/vistone/fingerprint/modules/profiles v0.0.0
)

replace (
	github.com/vistone/fingerprint/modules/generator => ../../modules/generator
	github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles
)
