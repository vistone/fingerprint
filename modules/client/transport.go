// Package client provides complete browser fingerprint simulationtransport layer
// uses unified fhttp type，supports automatic HTTP/2 → HTTP/1.1 fallback
package client

import (
	"bytes"
	"context"
	stdtls "crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"syscall"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
	"github.com/vistone/fingerprint/modules/profiles"
	"golang.org/x/sys/unix"
)

// SmartTransport smart transport layer，uses unified fhttp type
type SmartTransport struct {
	profile profiles.ClientProfile
	dialer  *net.Dialer
	// strictFingerprint disallow standard TLS compatibility fallback，ensure fingerprint chain path。
	strictFingerprint bool

	mu                sync.RWMutex
	hostProtocolCache map[string]string

	http2Transport *http2.Transport
}

// SetStrictFingerprint set strict fingerprint mode。
func (st *SmartTransport) SetStrictFingerprint(strict bool) {
	st.strictFingerprint = strict
}

// NewSmartTransport create smart transport layer
func NewSmartTransport(profile profiles.ClientProfile) (*SmartTransport, error) {
	st := &SmartTransport{
		profile:           profile,
		hostProtocolCache: make(map[string]string),
	}

	if err := st.configureTCP(); err != nil {
		return nil, err
	}

	st.initHTTP2()

	return st, nil
}

// configureTCP configure TCP/IP
func (st *SmartTransport) configureTCP() error {
	tcpip := st.profile.TCPIP
	if tcpip == nil {
		tcpip = &profiles.TCPIPFingerprint{
			IPVersion:     4,
			TTL:           128,
			WindowSize:    64240,
			MSS:           1460,
			WindowScale:   8,
			SAckPermitted: true,
			Timestamps:    true,
		}
	}

	st.dialer = &net.Dialer{
		Timeout:   TimeoutDialConnect,
		KeepAlive: KeepAliveInterval,
		Control: func(network, address string, c syscall.RawConn) error {
			var sockErr error
			err := c.Control(func(fd uintptr) {
				// prioritize applying core TCP/IP parameters，unsupported platform options will be safely ignored。
				if tcpip.TTL > 0 {
					sockErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_TTL, int(tcpip.TTL))
					if sockErr != nil {
						return
					}
				}

				if tcpip.WindowSize > 0 {
					_ = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_RCVBUF, int(tcpip.WindowSize))
					_ = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_SNDBUF, int(tcpip.WindowSize))
					_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_WINDOW_CLAMP, int(tcpip.WindowSize))
				}

				sockErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_NODELAY, 1)
				if sockErr != nil {
					return
				}

				if tcpip.MSS > 0 {
					_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_MAXSEG, int(tcpip.MSS))
				}

				if tcpip.Timestamps {
					_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_TIMESTAMP, 1)
				}

				_ = tcpip.SAckPermitted

				_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_QUICKACK, 1)
				_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, 7200)

				// JA4T current field used for profile consistency verification, reserved for subsequent wire-level alignment.
				_ = tcpip.JA4T
			})
			if err != nil {
				return err
			}
			return sockErr
		},
	}

	return nil
}

// initHTTP2 initialize HTTP/2
func (st *SmartTransport) initHTTP2() {
	st.http2Transport = &http2.Transport{
		DialTLS: st.dialTLS,
	}
}

// dialTLS establish TLS connection (for HTTP/2 use)
func (st *SmartTransport) dialTLS(network, addr string, cfg *tls.Config) (net.Conn, error) {
	// establish TCP connection
	tcpConn, err := st.dialer.Dial(network, addr)
	if err != nil {
		return nil, fmt.Errorf("dial TCP failed: %w", err)
	}

	// create uTLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
	}

	// apply fingerprint TLS configuration
	if st.profile.TLSVersion != 0 {
		tlsConfig.MinVersion = convertTLSVersion(st.profile.TLSVersion)
		tlsConfig.MaxVersion = convertTLSVersion(st.profile.TLSVersion)
	}
	if len(st.profile.CipherSuites) > 0 {
		tlsConfig.CipherSuites = convertCipherSuites(st.profile.CipherSuites)
	}

	// merge external configuration
	if cfg != nil && cfg.ServerName != "" {
		tlsConfig.ServerName = cfg.ServerName
	}

	// set ServerName
	if tlsConfig.ServerName == "" {
		host, _, _ := net.SplitHostPort(addr)
		if host != "" {
			tlsConfig.ServerName = host
		}
	}

	// use fingerprint ClientHello
	spec, err := st.getProfileClientHelloSpec()
	if err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("resolve profile ClientHello spec failed: %w", err)
	}

	clientHelloID := getClientHelloID(string(st.profile.BrowserType))
	if spec != nil {
		clientHelloID = tls.HelloCustom
	}
	tlsConn := tls.UClient(tcpConn, tlsConfig, clientHelloID, false, false, false)
	if spec != nil {
		if err := tlsConn.ApplyPreset(spec); err != nil {
			tcpConn.Close()
			return nil, fmt.Errorf("apply profile ClientHello spec failed: %w", err)
		}
	}

	// perform handshake
	if err := tlsConn.Handshake(); err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	return tlsConn, nil
}

// RoundTrip execute request (unified return *fhttp.Response)
func (st *SmartTransport) RoundTrip(req *fhttp.Request) (*fhttp.Response, error) {
	ctx := req.Context()
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}

	// check cached protocol preference
	st.mu.RLock()
	cachedProto := st.hostProtocolCache[host]
	st.mu.RUnlock()

	// prefer cached protocol
	if cachedProto == "http/1.1" {
		resp, err := st.roundTripHTTP1(ctx, req)
		if err == nil && !shouldRetryWithHTTP1(resp) {
			return resp, nil
		}
		if st.strictFingerprint {
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
		return st.roundTripHTTP1Compat(ctx, req)
	}

	// 1. attempt HTTP/2
	resp, err := st.roundTripHTTP2(ctx, req)
	if err == nil {
		// some sites/CDN may directly return 400/421 for current H2 fingerprint, fallback to HTTP/1.1 can recover.
		if shouldRetryWithHTTP1(resp) {
			h1Resp, h1Err := st.roundTripHTTP1(ctx, req)
			if h1Err == nil && !shouldRetryWithHTTP1(h1Resp) {
				if resp.Body != nil {
					_ = resp.Body.Close()
				}
				st.mu.Lock()
				st.hostProtocolCache[host] = "http/1.1"
				st.mu.Unlock()
				return h1Resp, nil
			}

			// final compatibility fallback: use standard TLS HTTP/1.1.
			if !st.strictFingerprint {
				compatResp, compatErr := st.roundTripHTTP1Compat(ctx, req)
				if compatErr == nil {
					if resp.Body != nil {
						_ = resp.Body.Close()
					}
					if h1Resp != nil && h1Resp.Body != nil {
						_ = h1Resp.Body.Close()
					}
					st.mu.Lock()
					st.hostProtocolCache[host] = "http/1.1"
					st.mu.Unlock()
					return compatResp, nil
				}
			}
		}

		st.mu.Lock()
		st.hostProtocolCache[host] = "h2"
		st.mu.Unlock()
		return resp, nil
	}

	// classify error information
	errType := classifyError(err)

	// for protocol error, TLS error, or network connection error, fallback to HTTP/1.1
	// but for context error or other clear errors, directly return
	if errType == ErrorTypeTimeout || errType == ErrorTypeCanceled {
		// timeout or cancellation error do not fallback
		return nil, err
	}

	// 2. fallback to HTTP/1.1
	resp, err = st.roundTripHTTP1(ctx, req)
	if err == nil {
		st.mu.Lock()
		st.hostProtocolCache[host] = "http/1.1"
		st.mu.Unlock()
		return resp, err
	}

	// 3. final compatibility fallback: handle ALPN negotiation/HTTP2 frame misidentification as HTTP/1.1 etc.
	if !st.strictFingerprint && shouldFallbackToHTTP1Compat(err) {
		compatResp, compatErr := st.roundTripHTTP1Compat(ctx, req)
		if compatErr == nil {
			st.mu.Lock()
			st.hostProtocolCache[host] = "http/1.1"
			st.mu.Unlock()
			return compatResp, nil
		}
	}

	// if both protocols fail, return first error (HTTP/2 error)
	return nil, err
}

// roundTripHTTP2 uses HTTP/2
func (st *SmartTransport) roundTripHTTP2(ctx context.Context, req *fhttp.Request) (*fhttp.Response, error) {
	return st.http2Transport.RoundTrip(req)
}

// roundTripHTTP1 uses HTTP/1.1
func (st *SmartTransport) roundTripHTTP1(ctx context.Context, req *fhttp.Request) (*fhttp.Response, error) {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}

	// create custom HTTP/1.1 client
	client := &http.Client{
		Transport: &http.Transport{
			Dial:              st.dialer.Dial,
			ForceAttemptHTTP2: false,
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return st.dialTLSForHTTP1(addr, host)
			},
		},
		Timeout: TimeoutRequest,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// convert to standard library request
	stdReq, err := fhttpToStdRequest(req)
	if err != nil {
		return nil, err
	}
	stdReq = stdReq.WithContext(ctx)

	// execute request
	stdResp, err := client.Do(stdReq)
	if err != nil {
		return nil, err
	}

	// convert back to fhttp response
	return stdResponseToFhttp(stdResp, req), nil
}

// roundTripHTTP1Compat uses standard TLS to execute HTTP/1.1 as final compatibility fallback path.
func (st *SmartTransport) roundTripHTTP1Compat(ctx context.Context, req *fhttp.Request) (*fhttp.Response, error) {
	stdReq, err := fhttpToStdRequest(req)
	if err != nil {
		return nil, err
	}
	stdReq = stdReq.WithContext(ctx)

	tr := &http.Transport{
		DialContext:       st.dialer.DialContext,
		ForceAttemptHTTP2: false,
		TLSClientConfig: &stdtls.Config{
			InsecureSkipVerify: true,
			MinVersion:         stdtls.VersionTLS12,
			NextProtos:         []string{"http/1.1"},
		},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   TimeoutRequest,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(stdReq)
	if err != nil {
		return nil, err
	}
	return stdResponseToFhttp(resp, req), nil
}

// fhttpToStdRequest convert fhttp.Request to standard http.Request
func fhttpToStdRequest(req *fhttp.Request) (*http.Request, error) {
	// read body
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	// create new request
	method := req.Method
	if method == "" {
		method = "GET"
	}

	urlStr := req.URL.String()
	if urlStr == "" {
		urlStr = "https://" + req.Host
	}

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	stdReq, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, err
	}

	// copy headers
	for key, values := range req.Header {
		for _, v := range values {
			stdReq.Header.Add(key, v)
		}
	}

	// set Host
	stdReq.Host = req.Host

	return stdReq, nil
}

// stdResponseToFhttp convert standard http.Response to fhttp.Response
func stdResponseToFhttp(resp *http.Response, req *fhttp.Request) *fhttp.Response {
	fresp := &fhttp.Response{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Proto:         resp.Proto,
		ProtoMajor:    resp.ProtoMajor,
		ProtoMinor:    resp.ProtoMinor,
		Header:        make(fhttp.Header),
		Body:          resp.Body,
		ContentLength: resp.ContentLength,
		Request:       req,
	}

	// copy headers
	for key, values := range resp.Header {
		for _, v := range values {
			fresp.Header.Add(key, v)
		}
	}

	return fresp
}

// dialTLSForHTTP1 establish TLS connection for HTTP/1.1
func (st *SmartTransport) dialTLSForHTTP1(addr, host string) (net.Conn, error) {
	tcpConn, err := st.dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// enforce HTTP/1.1 TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
		NextProtos:         []string{"http/1.1"},
	}

	if st.profile.TLSVersion != 0 {
		tlsConfig.MinVersion = convertTLSVersion(st.profile.TLSVersion)
		tlsConfig.MaxVersion = convertTLSVersion(st.profile.TLSVersion)
	}
	if len(st.profile.CipherSuites) > 0 {
		tlsConfig.CipherSuites = convertCipherSuites(st.profile.CipherSuites)
	}

	clientHelloID := getClientHelloID(string(st.profile.BrowserType))
	spec, err := st.getProfileClientHelloSpec()
	if err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("resolve profile ClientHello spec failed: %w", err)
	}
	if spec != nil {
		clientHelloID = tls.HelloCustom
	}
	tlsConn := tls.UClient(tcpConn, tlsConfig, clientHelloID, false, false, false)
	if spec != nil {
		if err := tlsConn.ApplyPreset(spec); err != nil {
			tcpConn.Close()
			return nil, fmt.Errorf("apply profile ClientHello spec failed: %w", err)
		}
	}

	if err := tlsConn.Handshake(); err != nil {
		tcpConn.Close()
		return nil, err
	}

	return tlsConn, nil
}

// Close closetransport layer
