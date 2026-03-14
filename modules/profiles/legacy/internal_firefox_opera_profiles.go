//nolint:composites
package profiles

import (
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var Firefox_117 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:               "Firefox",
		RandomExtensionOrder: false,
		Version:              "117",
		Seed:                 nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
				CompressionMethods: []byte{
					tls.CompressionNone,
				},
				Extensions: []tls.TLSExtension{
					&tls.SNIExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
						tls.CurveP521,
						tls.FAKEFFDHE2048,
						tls.FAKEFFDHE3072,
					}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},

					&tls.SessionTicketExtension{},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.StatusRequestExtension{},
					&tls.DelegatedCredentialsExtension{
						SupportedSignatureAlgorithms: []tls.SignatureScheme{
							tls.ECDSAWithP256AndSHA256,
							tls.ECDSAWithP384AndSHA384,
							tls.ECDSAWithP521AndSHA512,
							tls.ECDSAWithSHA1,
						},
					},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.KeyShareExtension{KeyShares: []tls.KeyShare{
						{Group: tls.X25519},
						{Group: tls.CurveP256},
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedVersionsExtension{Versions: []uint16{
						tls.VersionTLS13,
						tls.VersionTLS12,
					}},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.ECDSAWithP256AndSHA256,
						tls.ECDSAWithP384AndSHA384,
						tls.ECDSAWithP521AndSHA512,
						tls.PSSWithSHA256,
						tls.PSSWithSHA384,
						tls.PSSWithSHA512,
						tls.PKCS1WithSHA256,
						tls.PKCS1WithSHA384,
						tls.PKCS1WithSHA512,
						tls.ECDSAWithSHA1,
						tls.PKCS1WithSHA1,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.PSKKeyExchangeModesExtension{Modes: []uint8{
						tls.PskModeDHE,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.FakeRecordSizeLimitExtension{Limit: 0x4001},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingInitialWindowSize: 131072,
		http2.SettingMaxFrameSize:      16384,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 12517377,
	headerPriority: &http2.PriorityParam{
		StreamDep: 13,
		Exclusive: false,
		Weight:    41,
	},
	priorities: []http2.Priority{
		{StreamID: 3, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    200,
		}},
		{StreamID: 5, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    100,
		}},
		{StreamID: 7, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 9, PriorityParam: http2.PriorityParam{
			StreamDep: 7,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 11, PriorityParam: http2.PriorityParam{
			StreamDep: 3,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 13, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    240,
		}},
	},
}

var Firefox_110 = ClientProfile{
	clientHelloId: tls.HelloFirefox_110,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingInitialWindowSize: 131072,
		http2.SettingMaxFrameSize:      16384,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 12517377,
	headerPriority: &http2.PriorityParam{
		StreamDep: 13,
		Exclusive: false,
		Weight:    41,
	},
	priorities: []http2.Priority{
		{StreamID: 3, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    200,
		}},
		{StreamID: 5, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    100,
		}},
		{StreamID: 7, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 9, PriorityParam: http2.PriorityParam{
			StreamDep: 7,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 11, PriorityParam: http2.PriorityParam{
			StreamDep: 3,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 13, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    240,
		}},
	},
}

var Firefox_108 = ClientProfile{
	clientHelloId: tls.HelloFirefox_108,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingInitialWindowSize: 131072,
		http2.SettingMaxFrameSize:      16384,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 12517377,
	headerPriority: &http2.PriorityParam{
		StreamDep: 13,
		Exclusive: false,
		Weight:    41,
	},
	priorities: []http2.Priority{
		{StreamID: 3, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    200,
		}},
		{StreamID: 5, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    100,
		}},
		{StreamID: 7, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 9, PriorityParam: http2.PriorityParam{
			StreamDep: 7,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 11, PriorityParam: http2.PriorityParam{
			StreamDep: 3,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 13, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    240,
		}},
	},
}

var Firefox_106 = ClientProfile{
	clientHelloId: tls.HelloFirefox_106,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingInitialWindowSize: 131072,
		http2.SettingMaxFrameSize:      16384,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 12517377,
	headerPriority: &http2.PriorityParam{
		StreamDep: 13,
		Exclusive: false,
		Weight:    41,
	},
	priorities: []http2.Priority{
		{StreamID: 3, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    200,
		}},
		{StreamID: 5, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    100,
		}},
		{StreamID: 7, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 9, PriorityParam: http2.PriorityParam{
			StreamDep: 7,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 11, PriorityParam: http2.PriorityParam{
			StreamDep: 3,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 13, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    240,
		}},
	},
}

var Firefox_105 = ClientProfile{
	clientHelloId: tls.HelloFirefox_105,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingInitialWindowSize: 131072,
		http2.SettingMaxFrameSize:      16384,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 12517377,
	headerPriority: &http2.PriorityParam{
		StreamDep: 13,
		Exclusive: false,
		Weight:    41,
	},
	priorities: []http2.Priority{
		{StreamID: 3, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    200,
		}},
		{StreamID: 5, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    100,
		}},
		{StreamID: 7, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 9, PriorityParam: http2.PriorityParam{
			StreamDep: 7,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 11, PriorityParam: http2.PriorityParam{
			StreamDep: 3,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 13, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    240,
		}},
	},
}

var Firefox_104 = ClientProfile{
	clientHelloId: tls.HelloFirefox_104,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingInitialWindowSize: 131072,
		http2.SettingMaxFrameSize:      16384,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 12517377,
	headerPriority: &http2.PriorityParam{
		StreamDep: 13,
		Exclusive: false,
		Weight:    41,
	},
	priorities: []http2.Priority{
		{StreamID: 3, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    200,
		}},
		{StreamID: 5, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    100,
		}},
		{StreamID: 7, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 9, PriorityParam: http2.PriorityParam{
			StreamDep: 7,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 11, PriorityParam: http2.PriorityParam{
			StreamDep: 3,
			Exclusive: false,
			Weight:    0,
		}},
		{StreamID: 13, PriorityParam: http2.PriorityParam{
			StreamDep: 0,
			Exclusive: false,
			Weight:    240,
		}},
	},
}
