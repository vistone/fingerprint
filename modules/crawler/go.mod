module github.com/vistone/fingerprint/modules/crawler

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/ml v1.0.21
	github.com/vistone/fingerprint/modules/profiles v1.0.21
)

require (
	github.com/vistone/fingerprint/modules/core v1.0.21 // indirect
	github.com/vistone/fingerprint/modules/errors v1.0.20 // indirect
)

replace (
	github.com/vistone/fingerprint/modules/client => ../client
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/ml => ../ml
	github.com/vistone/fingerprint/modules/profiles => ../profiles
)
