module crawler-example

go 1.25.7

require github.com/vistone/fingerprint/modules/crawler v0.0.0

require (
	github.com/vistone/fingerprint/modules/core v1.0.21 // indirect
	github.com/vistone/fingerprint/modules/errors v1.0.20 // indirect
	github.com/vistone/fingerprint/modules/profiles v1.0.21 // indirect
)

replace github.com/vistone/fingerprint/modules/crawler => ../../modules/crawler

replace github.com/vistone/fingerprint/modules/core => ../../modules/core

replace github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles
