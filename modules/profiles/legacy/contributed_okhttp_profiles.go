//nolint:composites
package profiles

import (
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var Okhttp4Android13 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "OkHttp4Android13",
		Version: "4.10.0",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return Okhttp4Android10.GetClientHelloSpec()
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize: 16777216,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
	},
	headerPriority: &http2.PriorityParam{},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 16711681,
}
var Okhttp4Android12 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "OkHttp4Android12",
		Version: "4.10.0",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return Okhttp4Android10.GetClientHelloSpec()
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize: 16777216,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
	},
	headerPriority: &http2.PriorityParam{},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 16711681,
}

var Okhttp4Android11 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "OkHttp4Android11",
		Version: "4.10.0",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return Okhttp4Android10.GetClientHelloSpec()
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize: 16777216,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
	},
	headerPriority: &http2.PriorityParam{},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 16711681,
}

var Okhttp4Android10 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "OkHttp4Android10",
		Version: "4.10.0",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
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
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateNever},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
					}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},
					&tls.SessionTicketExtension{},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.StatusRequestExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.ECDSAWithP256AndSHA256,
						tls.PSSWithSHA256,
						tls.PKCS1WithSHA256,
						tls.ECDSAWithP384AndSHA384,
						tls.PSSWithSHA384,
						tls.PKCS1WithSHA384,
						tls.PSSWithSHA512,
						tls.PKCS1WithSHA512,
						tls.PKCS1WithSHA1,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.KeyShareExtension{KeyShares: []tls.KeyShare{
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
						tls.VersionTLS13,
						tls.VersionTLS12,
					}},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize: 16777216,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
	},
	headerPriority: &http2.PriorityParam{},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 16711681,
}

var Okhttp4Android9 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "OkHttp4Android9",
		Version: "4.10.0",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
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
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateNever},
					&tls.SNIExtension{},
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
						tls.PKCS1WithSHA1,
					}},
					&tls.StatusRequestExtension{},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
					}},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize: 16777216,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
	},
	headerPriority: &http2.PriorityParam{},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 16711681,
}

var Okhttp4Android8 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "OkHttp4Android8",
		Version: "4.10.0",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
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
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateNever},
					&tls.SNIExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.SessionTicketExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.ECDSAWithP256AndSHA256,
						tls.PKCS1WithSHA256,
						tls.ECDSAWithP384AndSHA384,
						tls.PKCS1WithSHA384,
						tls.ECDSAWithP521AndSHA512,
						tls.PKCS1WithSHA512,
						tls.PKCS1WithSHA1,
					}},
					&tls.StatusRequestExtension{},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.X25519,
						tls.CurveP256,
						tls.CurveP384,
					}},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize: 16777216,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
	},
	headerPriority: &http2.PriorityParam{},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 16711681,
}

var Okhttp4Android7 = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "OkHttp4Android7",
		Version: "4.10.0",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
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
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateNever},
					&tls.SNIExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.SessionTicketExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.PKCS1WithSHA512,
						tls.ECDSAWithP521AndSHA512,
						tls.PKCS1WithSHA384,
						tls.ECDSAWithP384AndSHA384,
						tls.PKCS1WithSHA256,
						tls.ECDSAWithP256AndSHA256,
						0x0301,
						0x0303,
						tls.PKCS1WithSHA1,
						tls.ECDSAWithSHA1,
					}},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.SupportedPointsExtension{SupportedPoints: []byte{
						tls.PointFormatUncompressed,
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.CurveP256,
						tls.CurveP384,
						tls.CurveP521,
					}},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingInitialWindowSize: 16777216,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingInitialWindowSize,
	},
	headerPriority: &http2.PriorityParam{},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 16711681,
}
