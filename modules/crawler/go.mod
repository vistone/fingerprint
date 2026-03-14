module github.com/vistone/fingerprint/modules/crawler

go 1.25.7

require (
	github.com/vistone/fingerprint/modules/core v1.0.21
	github.com/vistone/fingerprint/modules/profiles v1.0.21
	github.com/vistone/fingerprint/modules/client v1.0.21
	github.com/chromedp/chromedp v0.11.2
	github.com/bogdanfinn/fhttp v0.6.8
	github.com/bogdanfinn/utls v1.7.7-barnius
)

replace (
	github.com/vistone/fingerprint/modules/core => ../core
	github.com/vistone/fingerprint/modules/profiles => ../profiles
	github.com/vistone/fingerprint/modules/client => ../client
)