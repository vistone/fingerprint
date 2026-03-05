module github.com/vistone/fingerprint/examples/basic

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v0.0.0
	github.com/vistone/fingerprint/modules/profiles v0.0.0
	github.com/vistone/fingerprint/modules/tls v0.0.0
)

replace (
	github.com/vistone/fingerprint/modules/core => ../../modules/core
	github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles
	github.com/vistone/fingerprint/modules/tls => ../../modules/tls
)
