module github.com/vistone/fingerprint/modules/waf

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.28
	github.com/vistone/fingerprint/modules/defense v1.0.28
	github.com/vistone/fingerprint/modules/agent v1.0.28
	github.com/vistone/fingerprint/modules/ml v1.0.28
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/defense => ../defense
	github.com/vistone/fingerprint/modules/agent => ../agent
	github.com/vistone/fingerprint/modules/ml => ../ml
)