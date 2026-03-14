//nolint:composites
package profiles

import (
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var Safari_15_6_1 = ClientProfile{
	clientHelloId: tls.HelloSafari_15_6_1,
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize:    4194304,
		http2.SettingMaxConcurrentStreams: 100,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
		http2.SettingMaxConcurrentStreams,
	},
	pseudoHeaderOrder: []string{
		":method",
		":scheme",
		":path",
		":authority",
	},
	connectionFlow: 10485760,
}

var Safari_16_0 = ClientProfile{
	clientHelloId: tls.HelloSafari_16_0,
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize:    4194304,
		http2.SettingMaxConcurrentStreams: 100,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
		http2.SettingMaxConcurrentStreams,
	},
	pseudoHeaderOrder: []string{
		":method",
		":scheme",
		":path",
		":authority",
	},
	connectionFlow: 10485760,
}

var Safari_Ipad_15_6 = ClientProfile{
	clientHelloId: tls.HelloIPad_15_6,
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize:    2097152,
		http2.SettingMaxConcurrentStreams: 100,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
		http2.SettingMaxConcurrentStreams,
	},
	pseudoHeaderOrder: []string{
		":method",
		":scheme",
		":path",
		":authority",
	},
	connectionFlow: 10485760,
}
