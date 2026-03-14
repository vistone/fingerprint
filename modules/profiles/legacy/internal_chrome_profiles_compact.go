//nolint:composites
package profiles

import (
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var Chrome_120 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:               "Chrome",
		RandomExtensionOrder: false,
		Version:              "120",
		Seed:                 nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.GREASE_PLACEHOLDER,
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
				CompressionMethods: []uint8{
					tls.CompressionNone,
				},
				Extensions: []tls.TLSExtension{
					&tls.UtlsGREASEExtension{},
					&tls.SNIExtension{},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.PSKKeyExchangeModesExtension{Modes: []uint8{
						tls.PskModeDHE,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedVersionsExtension{Versions: []uint16{
						tls.GREASE_PLACEHOLDER,
						tls.VersionTLS13,
						tls.VersionTLS12,
					}},
					&tls.StatusRequestExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.SessionTicketExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.ECDSAWithP256AndSHA256,
						tls.PSSWithSHA256,
						tls.PKCS1WithSHA256,
						tls.ECDSAWithP384AndSHA384,
						tls.PSSWithSHA384,
						tls.PKCS1WithSHA384,
						tls.PSSWithSHA512,
						tls.PKCS1WithSHA512,
					}},
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					tls.BoringGREASEECH(),
					&tls.SCTExtension{},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.KeyShareExtension{KeyShares: []tls.KeyShare{
						{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
						{Group: tls.X25519},
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.GREASE_PLACEHOLDER,
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
					}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},
					&tls.ApplicationSettingsExtension{
						SupportedProtocols: []string{"h2"},
					},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{
						tls.CertCompressionBrotli,
					}},
					&tls.UtlsGREASEExtension{},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:   65536,
		http2.SettingEnablePush:        0,
		http2.SettingInitialWindowSize: 6291456,
		http2.SettingMaxHeaderListSize: 262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_112 = ClientProfile{
	clientHelloId: tls.HelloChrome_112,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_116_PSK = ClientProfile{
	clientHelloId: tls.HelloChrome_112_PSK,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_116_PSK_PQ = ClientProfile{
	clientHelloId: tls.HelloChrome_115_PQ_PSK,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_111 = ClientProfile{
	clientHelloId: tls.HelloChrome_111,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_110 = ClientProfile{
	clientHelloId: tls.HelloChrome_110,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_109 = ClientProfile{
	clientHelloId: tls.HelloChrome_109,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_108 = ClientProfile{
	clientHelloId: tls.HelloChrome_108,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_107 = ClientProfile{
	clientHelloId: tls.HelloChrome_107,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_106 = ClientProfile{
	clientHelloId: tls.HelloChrome_106,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_105 = ClientProfile{
	clientHelloId: tls.HelloChrome_105,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_104 = ClientProfile{
	clientHelloId: tls.HelloChrome_104,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}

var Chrome_103 = ClientProfile{
	clientHelloId: tls.HelloChrome_103,
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":authority",
		":scheme",
		":path",
	},
	connectionFlow: 15663105,
}
