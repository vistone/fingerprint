//nolint:composites
package profiles

import (
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var Safari_IOS_17_0 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:               "iOS",
		RandomExtensionOrder: false,
		Version:              "17.0",
		Seed:                 nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.GREASE_PLACEHOLDER,
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_3DES_EDE_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
					tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
				},
				CompressionMethods: []uint8{
					tls.CompressionNone,
				},
				Extensions: []tls.TLSExtension{
					&tls.UtlsGREASEExtension{},
					&tls.SNIExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.GREASE_PLACEHOLDER,
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
						tls.CurveP521,
					}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.StatusRequestExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.ECDSAWithP256AndSHA256,
						tls.PSSWithSHA256,
						tls.PKCS1WithSHA256,
						tls.ECDSAWithP384AndSHA384,
						tls.ECDSAWithSHA1,
						tls.PSSWithSHA384,
						tls.PSSWithSHA384,
						tls.PKCS1WithSHA384,
						tls.PSSWithSHA512,
						tls.PKCS1WithSHA512,
						tls.PKCS1WithSHA1,
					}},
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
						tls.VersionTLS11,
						tls.VersionTLS10,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{
						tls.CertCompressionZlib,
					}},
					&tls.UtlsGREASEExtension{},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingEnablePush:           0,
		http2.SettingInitialWindowSize:    2097152,
		http2.SettingMaxConcurrentStreams: 100,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingEnablePush,
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

var Safari_IOS_18_5 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:               "iOS",
		RandomExtensionOrder: false,
		Version:              "18.5",
		Seed:                 nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.GREASE_PLACEHOLDER,
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_3DES_EDE_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
					tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
				},
				CompressionMethods: []uint8{
					tls.CompressionNone,
				},
				Extensions: []tls.TLSExtension{
					&tls.UtlsGREASEExtension{},
					&tls.SNIExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.GREASE_PLACEHOLDER,
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
						tls.CurveP521,
					}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.StatusRequestExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.ECDSAWithP256AndSHA256,
						tls.PSSWithSHA256,
						tls.PKCS1WithSHA256,
						tls.ECDSAWithP384AndSHA384,
						tls.PSSWithSHA384,
						tls.PSSWithSHA384,
						tls.PKCS1WithSHA384,
						tls.PSSWithSHA512,
						tls.PKCS1WithSHA512,
						tls.PKCS1WithSHA1,
					}},
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
						tls.VersionTLS11,
						tls.VersionTLS10,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{
						tls.CertCompressionZlib,
					}},
					&tls.UtlsGREASEExtension{},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 100,
		http2.SettingInitialWindowSize:    2097152,
		http2.SettingNoRFC7540Priorities:  1,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingNoRFC7540Priorities,
	},
	pseudoHeaderOrder: []string{
		":method",
		":scheme",
		":authority",
		":path",
	},
	connectionFlow: 10420225,
	headerPriority: &http2.PriorityParam{
		StreamDep: 0,
		Exclusive: false,
		Weight:    255,
	},
}

var Safari_IOS_18_0 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:               "iOS",
		RandomExtensionOrder: false,
		Version:              "18.0",
		Seed:                 nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.GREASE_PLACEHOLDER,
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_3DES_EDE_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
					tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
				},
				CompressionMethods: []uint8{
					tls.CompressionNone,
				},
				Extensions: []tls.TLSExtension{
					&tls.UtlsGREASEExtension{},
					&tls.SNIExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.GREASE_PLACEHOLDER,
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
						tls.CurveP521,
					}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.StatusRequestExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.ECDSAWithP256AndSHA256,
						tls.PSSWithSHA256,
						tls.PKCS1WithSHA256,
						tls.ECDSAWithP384AndSHA384,
						tls.ECDSAWithSHA1,
						tls.PSSWithSHA384,
						tls.PSSWithSHA384,
						tls.PKCS1WithSHA384,
						tls.PSSWithSHA512,
						tls.PKCS1WithSHA512,
						tls.PKCS1WithSHA1,
					}},
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
						tls.VersionTLS11,
						tls.VersionTLS10,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{
						tls.CertCompressionZlib,
					}},
					&tls.UtlsGREASEExtension{},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingEnablePush:           0,
		http2.SettingMaxConcurrentStreams: 100,
		http2.SettingInitialWindowSize:    2097152,
		0x8:                               1,
		0x9:                               1,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		0x8,
		0x9,
	},
	pseudoHeaderOrder: []string{
		":method",
		":scheme",
		":authority",
		":path",
	},
	connectionFlow: 10420225,
}

var Safari_IOS_16_0 = ClientProfile{
	clientHelloId: tls.HelloIOS_16_0,
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

var Safari_IOS_15_5 = ClientProfile{
	clientHelloId: tls.HelloIOS_15_5,
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

var Safari_IOS_15_6 = ClientProfile{
	clientHelloId: tls.HelloIOS_15_6,
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
