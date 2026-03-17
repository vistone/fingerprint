package profiles

import "github.com/vistone/fingerprint/modules/core"

func ensureChromiumCommonDefaults(p *ClientProfile) {
	ensureChromiumHTTP2Defaults(p)
	ensureChromiumHTTP3Defaults(p)
	ensureChromiumHeaderDefaults(p)
}

func ensureChromiumHTTP2Defaults(p *ClientProfile) {
	if p.HTTP2Settings.HeaderTableSize == 0 && p.HTTP2Settings.InitialWindowSize == 0 {
		p.HTTP2Settings = core.HTTP2Settings{
			HeaderTableSize:      65536,
			EnablePush:           0,
			MaxConcurrentStreams: 1000,
			InitialWindowSize:    6291456,
			MaxFrameSize:         16384,
			MaxHeaderListSize:    262144,
		}
		p.PseudoHeaderOrder = []string{":method", ":authority", ":scheme", ":path"}
	}
	if p.ConnectionFlow == 0 {
		p.ConnectionFlow = 15663105
	}
}

func ensureChromiumHTTP3Defaults(p *ClientProfile) {
	if p.HTTP3Settings != nil {
		return
	}
	p.HTTP3Settings = &core.HTTP3Settings{
		QUICVersion:            core.QUICVersion1,
		InitialMaxData:         16777216,
		InitialMaxStreamData:   6291456,
		InitialMaxStreamsBidi:  100,
		InitialMaxStreamsUni:   100,
		MaxUDPPayloadSize:      1472,
		AckDelayExponent:       3,
		MaxAckDelay:            25,
		DisableActiveMigration: false,
	}
	p.QUICVersions = []uint32{core.QUICVersion1}
}

func ensureChromiumHeaderDefaults(p *ClientProfile) {
	if p.Headers == nil {
		p.Headers = &core.HTTPHeaders{}
	}
	h := p.Headers
	if h.Accept == "" {
		h.Accept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	}
	if h.AcceptLanguage == "" {
		h.AcceptLanguage = "en-US,en;q=0.9"
	}
	if h.AcceptEncoding == "" {
		h.AcceptEncoding = "gzip, deflate, br"
	}
	if h.SecFetchSite == "" {
		h.SecFetchSite = "none"
	}
	if h.SecFetchMode == "" {
		h.SecFetchMode = "navigate"
	}
	if h.SecFetchDest == "" {
		h.SecFetchDest = "document"
	}
	if h.UpgradeInsecureRequests == "" {
		h.UpgradeInsecureRequests = "1"
	}
}
