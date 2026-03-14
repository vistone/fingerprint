module crawler-waf-integration

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/crawler v0.0.0
	github.com/vistone/fingerprint/modules/waf v0.0.0
)

replace github.com/vistone/fingerprint/modules/crawler => ../../modules/crawler

replace github.com/vistone/fingerprint/modules/waf => ../../modules/waf

replace github.com/vistone/fingerprint/modules/core => ../../modules/core

replace github.com/vistone/fingerprint/modules/profiles => ../../modules/profiles

replace github.com/vistone/fingerprint/modules/defense => ../../modules/defense

replace github.com/vistone/fingerprint/modules/ml => ../../modules/ml

replace github.com/vistone/fingerprint/modules/agent => ../../modules/agent
