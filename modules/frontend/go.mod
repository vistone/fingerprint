module github.com/vistone/fingerprint/modules/frontend

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.3
	github.com/vistone/fingerprint/modules/ml v1.0.3
)

require github.com/vistone/fingerprint/modules/profiles v1.0.3 // indirect

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/ml => ../ml
)
