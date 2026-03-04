module example

go 1.25.7

require github.com/vistone/fingerprint v0.0.0

require (
	github.com/vistone/fingerprint/modules/core v0.0.0 // indirect
	github.com/vistone/fingerprint/modules/defense v0.0.0 // indirect
	github.com/vistone/fingerprint/modules/frontend v0.0.0 // indirect
	github.com/vistone/fingerprint/modules/gateway v0.0.0 // indirect
	github.com/vistone/fingerprint/modules/http v0.0.0 // indirect
	github.com/vistone/fingerprint/modules/ml v0.0.0 // indirect
	github.com/vistone/fingerprint/modules/profiles v0.0.0 // indirect
	github.com/vistone/fingerprint/modules/tls v0.0.0 // indirect
)

replace (
	github.com/vistone/fingerprint => ../..
	github.com/vistone/fingerprint/modules/core => ../../modules/core
	github.com/vistone/fingerprint/modules/defense => ../../modules/defense
	github.com/vistone/fingerprint/modules/frontend => ../../modules/frontend
	github.com/vistone/fingerprint/modules/gateway => ../../modules/gateway
	github.com/vistone/fingerprint/modules/http => ../../modules/http
	github.com/vistone/fingerprint/modules/ml => ../../modules/ml
	github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles
	github.com/vistone/fingerprint/modules/tls => ../../modules/tls
)
