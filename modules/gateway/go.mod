module github.com/vistone/fingerprint/modules/gateway

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v0.0.0
	github.com/vistone/fingerprint/modules/defense v0.0.0
	github.com/vistone/fingerprint/modules/frontend v0.0.0
	github.com/vistone/fingerprint/modules/ml v0.0.0
	github.com/vistone/fingerprint/modules/profiles v0.0.0
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/defense => ../defense
	github.com/vistone/fingerprint/modules/frontend => ../frontend
	github.com/vistone/fingerprint/modules/ml => ../ml
	github.com/vistone/fingerprint/modules/profiles => ../profiles
)
