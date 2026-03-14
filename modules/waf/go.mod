module github.com/vistone/fingerprint/modules/waf

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/agent v1.0.21
	github.com/vistone/fingerprint/modules/core v1.0.21
	github.com/vistone/fingerprint/modules/defense v1.0.21
	github.com/vistone/fingerprint/modules/ml v1.0.21
	github.com/vistone/fingerprint/modules/profiles v1.0.20
)

require github.com/vistone/fingerprint/modules/errors v1.0.20 // indirect

replace (
	github.com/vistone/fingerprint/modules/agent => ../agent
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/defense => ../defense
	github.com/vistone/fingerprint/modules/ml => ../ml
)
