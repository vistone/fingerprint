module github.com/vistone/fingerprint/examples/advanced

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.3
	github.com/vistone/fingerprint/modules/fingerprint v1.0.3
	github.com/vistone/fingerprint/modules/profiles v1.0.3
)

require (
	github.com/vistone/fingerprint/modules/defense v1.0.3 // indirect
	github.com/vistone/fingerprint/modules/frontend v1.0.3 // indirect
	github.com/vistone/fingerprint/modules/gateway v1.0.3 // indirect
	github.com/vistone/fingerprint/modules/http v1.0.3 // indirect
	github.com/vistone/fingerprint/modules/ml v1.0.3 // indirect
	github.com/vistone/fingerprint/modules/tls v1.0.3 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.79.1 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace (
	github.com/vistone/fingerprint/modules/core => ../../modules/core
	github.com/vistone/fingerprint/modules/fingerprint => ../../modules/fingerprint
	github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles
)
