//nolint:composites
package profiles

import (
	"math"

	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var NikeAndroidMobile = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "NikeAndroidCustom",
		Version: "1",
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
				CompressionMethods: []uint8{
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
						tls.VersionTLS13,
						tls.VersionTLS12,
					}},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	},
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      4096,
		http2.SettingMaxConcurrentStreams: math.MaxUint32,
		http2.SettingInitialWindowSize:    16777216,
		http2.SettingMaxFrameSize:         16384,
		http2.SettingMaxHeaderListSize:    math.MaxUint32,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 15663105,
}

var CloudflareCustom = ClientProfile{
	clientHelloId: tls.ClientHelloID{
		Client:  "CloudflareCustom",
		Version: "1",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.FAKE_TLS_EMPTY_RENEGOTIATION_INFO_SCSV,
				},
				CompressionMethods: []uint8{
					tls.CompressionNone,
				},
				Extensions: []tls.TLSExtension{
					&tls.SNIExtension{},
					&tls.SupportedPointsExtension{SupportedPoints: []uint8{
						tls.PointFormatUncompressed,
						1, // ansiX962_compressed_prime
						2, // ansiX962_compressed_char2
					}},
					//nolint:composites

					//nolint:composites

					//nolint:composites

					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.CurveID(0x0017),
					}},
					&tls.SessionTicketExtension{},
					// due to that we do not care about http2 frame settings
					&tls.ALPNExtension{AlpnProtocols: []string{"http/1.1"}},
					&tls.GenericExtension{Id: 22}, // encrypt_then_mac
					&tls.ExtendedMasterSecretExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						tls.ECDSAWithP256AndSHA256,
						tls.ECDSAWithP384AndSHA384,
						tls.ECDSAWithP521AndSHA512,
						0x0807,
						0x0808,
						0x0809,
						0x080a,
						0x080b,
						tls.PSSWithSHA256,
						tls.PSSWithSHA384,
						tls.PSSWithSHA512,
						tls.PKCS1WithSHA256,
						tls.PKCS1WithSHA384,
						tls.PKCS1WithSHA512,
						0x0303,
						0x0203,
						0x0301,
						0x0201,
						0x0302,
						0x0202,
						0x0402,
						0x0502,
						0x0602,
					}},
				},
			}, nil
		},
	},

	//actually the h2 settings are not relevant, because this client does only support http1
	settings: map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      4096,
		http2.SettingMaxConcurrentStreams: math.MaxUint32,
		http2.SettingInitialWindowSize:    16777216,
		http2.SettingMaxFrameSize:         16384,
		http2.SettingMaxHeaderListSize:    math.MaxUint32,
	},
	settingsOrder: []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
		http2.SettingMaxHeaderListSize,
	},
	pseudoHeaderOrder: []string{
		":method",
		":path",
		":authority",
		":scheme",
	},
	connectionFlow: 15663105,
}
