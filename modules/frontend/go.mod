module github.com/vistone/fingerprint/modules/frontend

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.9
	github.com/vistone/fingerprint/modules/ml v1.0.9
)

require github.com/vistone/fingerprint/modules/profiles v1.0.9

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/ml => ../ml
)
