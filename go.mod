module github.com/vistone/fingerprint

go 1.25.4

require (
	github.com/bogdanfinn/fhttp v0.6.3
	github.com/bogdanfinn/utls v1.7.4-barnius
	github.com/vistone/domaindns v1.0.0
	github.com/vistone/localippool v1.0.0
	github.com/vistone/logs v1.0.0
	github.com/vistone/netconnpool v1.0.1
	github.com/vistone/quic v1.0.0
)

replace (
	github.com/vistone/domaindns => ../domaindns
	github.com/vistone/localippool => ../localippool
	github.com/vistone/logs => ../logs
	github.com/vistone/netconnpool => ../netconnpool
	github.com/vistone/quic => ../quic
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/miekg/dns v1.1.68 // indirect
	github.com/quic-go/quic-go v0.57.1 // indirect
	github.com/vishvananda/netlink v1.3.1 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
