// Package client 提供完整的浏览器指纹模拟传输层
// 统一使用 fhttp 类型，支持 HTTP/2 → HTTP/1.1 自动回退
package client

import (
	"bytes"
	"context"
	stdtls "crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
	"github.com/vistone/fingerprint/modules/profiles"
	legacyprofiles "github.com/vistone/fingerprint/modules/profiles/legacy"
	"golang.org/x/sys/unix"
)

// SmartTransport 智能传输层，统一使用 fhttp 类型
type SmartTransport struct {
	profile profiles.ClientProfile
	dialer  *net.Dialer
	// strictFingerprint 禁止标准 TLS 兼容回退，确保走指纹链路。
	strictFingerprint bool

	mu                sync.RWMutex
	hostProtocolCache map[string]string

	http2Transport *http2.Transport
}

// SetStrictFingerprint 设置严格指纹模式。
func (st *SmartTransport) SetStrictFingerprint(strict bool) {
	st.strictFingerprint = strict
}

// NewSmartTransport 创建智能传输层
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

// configureTCP 配置 TCP/IP
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
				// 优先应用核心 TCP/IP 参数，不支持的平台选项将被安全忽略。
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

				// JA4T 当前用于 profile 一致性校验，保留字段以便后续 wire-level 对齐。
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

// initHTTP2 初始化 HTTP/2
func (st *SmartTransport) initHTTP2() {
	st.http2Transport = &http2.Transport{
		DialTLS: st.dialTLS,
	}
}

// dialTLS 建立 TLS 连接 (供 HTTP/2 使用)
func (st *SmartTransport) dialTLS(network, addr string, cfg *tls.Config) (net.Conn, error) {
	// 建立 TCP 连接
	tcpConn, err := st.dialer.Dial(network, addr)
	if err != nil {
		return nil, fmt.Errorf("dial TCP failed: %w", err)
	}

	// 创建 uTLS 配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
	}

	// 应用指纹 TLS 配置
	if st.profile.TLSVersion != 0 {
		tlsConfig.MinVersion = convertTLSVersion(st.profile.TLSVersion)
		tlsConfig.MaxVersion = convertTLSVersion(st.profile.TLSVersion)
	}
	if len(st.profile.CipherSuites) > 0 {
		tlsConfig.CipherSuites = convertCipherSuites(st.profile.CipherSuites)
	}

	// 合并外部配置
	if cfg != nil && cfg.ServerName != "" {
		tlsConfig.ServerName = cfg.ServerName
	}

	// 设置 ServerName
	if tlsConfig.ServerName == "" {
		host, _, _ := net.SplitHostPort(addr)
		if host != "" {
			tlsConfig.ServerName = host
		}
	}

	// 使用指纹 ClientHello
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

	// 执行握手
	if err := tlsConn.Handshake(); err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	return tlsConn, nil
}

// RoundTrip 执行请求 (统一返回 *fhttp.Response)
func (st *SmartTransport) RoundTrip(req *fhttp.Request) (*fhttp.Response, error) {
	ctx := req.Context()
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}

	// 检查缓存的协议偏好
	st.mu.RLock()
	cachedProto := st.hostProtocolCache[host]
	st.mu.RUnlock()

	// 优先使用缓存的协议
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

	// 1. 尝试 HTTP/2
	resp, err := st.roundTripHTTP2(ctx, req)
	if err == nil {
		// 某些站点/CDN 对当前 H2 指纹会直接返回 400/421，回退到 HTTP/1.1 可恢复。
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

			// 最后一层兼容回退：使用标准 TLS 的 HTTP/1.1。
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

	// 分类错误信息
	errType := classifyError(err)

	// 对于协议错误、TLS错误、或网络连接错误，回退到 HTTP/1.1
	// 但对于上下文错误或其他明确的错误，直接返回
	if errType == ErrorTypeTimeout || errType == ErrorTypeCanceled {
		// 超时或取消错误不回退
		return nil, err
	}

	// 2. 回退到 HTTP/1.1
	resp, err = st.roundTripHTTP1(ctx, req)
	if err == nil {
		st.mu.Lock()
		st.hostProtocolCache[host] = "http/1.1"
		st.mu.Unlock()
		return resp, err
	}

	// 3. 最终兼容回退：处理 ALPN 协商/HTTP2 帧误判为 HTTP/1.1 等问题。
	if !st.strictFingerprint && shouldFallbackToHTTP1Compat(err) {
		compatResp, compatErr := st.roundTripHTTP1Compat(ctx, req)
		if compatErr == nil {
			st.mu.Lock()
			st.hostProtocolCache[host] = "http/1.1"
			st.mu.Unlock()
			return compatResp, nil
		}
	}

	// 如果两个协议都失败，返回第一个错误（HTTP/2错误）
	return nil, err
}

// roundTripHTTP2 使用 HTTP/2
func (st *SmartTransport) roundTripHTTP2(ctx context.Context, req *fhttp.Request) (*fhttp.Response, error) {
	return st.http2Transport.RoundTrip(req)
}

// roundTripHTTP1 使用 HTTP/1.1
func (st *SmartTransport) roundTripHTTP1(ctx context.Context, req *fhttp.Request) (*fhttp.Response, error) {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}

	// 创建自定义的 HTTP/1.1 客户端
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

	// 转换为标准库请求
	stdReq, err := fhttpToStdRequest(req)
	if err != nil {
		return nil, err
	}
	stdReq = stdReq.WithContext(ctx)

	// 执行请求
	stdResp, err := client.Do(stdReq)
	if err != nil {
		return nil, err
	}

	// 转换回 fhttp 响应
	return stdResponseToFhttp(stdResp, req), nil
}

// roundTripHTTP1Compat 使用标准 TLS 执行 HTTP/1.1，作为最终兼容回退路径。
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

// fhttpToStdRequest 转换 fhttp.Request 到标准 http.Request
func fhttpToStdRequest(req *fhttp.Request) (*http.Request, error) {
	// 读取 body
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	// 创建新请求
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

	// 复制 headers
	for key, values := range req.Header {
		for _, v := range values {
			stdReq.Header.Add(key, v)
		}
	}

	// 设置 Host
	stdReq.Host = req.Host

	return stdReq, nil
}

// stdResponseToFhttp 转换标准 http.Response 到 fhttp.Response
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

	// 复制 headers
	for key, values := range resp.Header {
		for _, v := range values {
			fresp.Header.Add(key, v)
		}
	}

	return fresp
}

// dialTLSForHTTP1 为 HTTP/1.1 建立 TLS 连接
func (st *SmartTransport) dialTLSForHTTP1(addr, host string) (net.Conn, error) {
	tcpConn, err := st.dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// 强制 HTTP/1.1 的 TLS 配置
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

// Close 关闭传输层
func (st *SmartTransport) Close() error {
	return nil
}

// shouldFallbackToHTTP1Compat 判断是否应尝试标准 TLS 的 HTTP/1.1 兼容路径。
func shouldFallbackToHTTP1Compat(err error) bool {
	if err == nil {
		return false
	}

	errLower := strings.ToLower(err.Error())

	patterns := []string{
		"http/1.x transport connection broken",
		"malformed http response",
		"http2_handshake_failed",
		"first record does not look like a tls handshake",
		"tls handshake",
		"alpn",
	}

	for _, p := range patterns {
		if strings.Contains(errLower, p) {
			return true
		}
	}

	return false
}

// ErrorType 错误类型
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeProtocol
	ErrorTypeNetwork
	ErrorTypeTLS
	ErrorTypeTimeout
	ErrorTypeCanceled
	ErrorTypeDNS
)

// classifyError 分类错误
func classifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := err.Error()

	// 根据上下文错误判断
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorTypeTimeout
	}
	if errors.Is(err, context.Canceled) {
		return ErrorTypeCanceled
	}

	// 根据错误信息判断
	timeoutPatterns := []string{"timeout", "i/o timeout", "deadline exceeded"}
	for _, pattern := range timeoutPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeTimeout
		}
	}

	dnsPatterns := []string{"no such host", "lookup", "nameserver"}
	for _, pattern := range dnsPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeDNS
		}
	}

	protocolPatterns := []string{
		"PROTOCOL_ERROR",
		"http2:",
		"stream error",
		"broken pipe",
		"reset by peer",
		"connection reset",
	}
	for _, pattern := range protocolPatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return ErrorTypeProtocol
		}
	}

	tlsPatterns := []string{"tls", "certificate", "handshake"}
	for _, pattern := range tlsPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeTLS
		}
	}

	networkPatterns := []string{
		"connection refused",
		"connection aborted",
		"no route to host",
		"network is unreachable",
		"EOF",
	}
	for _, pattern := range networkPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeNetwork
		}
	}

	return ErrorTypeUnknown
}

// isProtocolError 检查是否为协议错误（已废弃，使用 classifyError）
func isProtocolError(err error) bool {
	if err == nil {
		return false
	}
	return classifyError(err) == ErrorTypeProtocol
}

func shouldRetryWithHTTP1(resp *fhttp.Response) bool {
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusMisdirectedRequest
}

// getProfileClientHelloSpec 基于 profile 解析可复用的细粒度 ClientHello 规范。
// 当不存在 legacy 映射时返回 nil，调用方将回落到现有 Auto ClientHello 行为。
func (st *SmartTransport) getProfileClientHelloSpec() (*tls.ClientHelloSpec, error) {
	legacyProfileID := resolveLegacyProfileID(st.profile)
	if legacyProfileID == "" {
		return nil, nil
	}

	legacyProfile, ok := legacyprofiles.GetClientProfile(legacyProfileID)
	if !ok {
		return nil, nil
	}

	spec, err := legacyProfile.GetClientHelloSpec()
	if err != nil {
		// 某些 legacy profile 尚未实现 ToSpec，遇到时保持兼容回退到 Auto ClientHello。
		return nil, nil
	}

	return &spec, nil
}

// resolveLegacyProfileID 返回可用于 uTLS ApplyPreset 的 legacy profile ID。
// 规则：优先精确 ID，其次按浏览器类型和主版本号就近匹配，最后退到该浏览器最新可用版本。
func resolveLegacyProfileID(profile profiles.ClientProfile) string {
	if profile.ID != "" {
		if _, ok := legacyprofiles.GetClientProfile(profile.ID); ok {
			return profile.ID
		}
	}

	browser := strings.ToLower(string(profile.BrowserType))
	if browser == "" {
		return ""
	}
	prefix := browser + "_"

	hasTargetMajor := false
	targetMajor := 0
	if profile.BrowserVersion != "" {
		if major, ok := parseLeadingMajor(profile.BrowserVersion); ok {
			targetMajor = major
			hasTargetMajor = true
		}
	}

	bestUnderOrEqualID := ""
	bestUnderOrEqualMajor := -1
	latestID := ""
	latestMajor := -1

	legacyIDs := legacyprofiles.GetAllProfiles()
	sort.Strings(legacyIDs)
	for _, id := range legacyIDs {
		idLower := strings.ToLower(id)
		if !strings.HasPrefix(idLower, prefix) {
			continue
		}
		if browser == "safari" && (strings.HasPrefix(idLower, "safari_ios_") || strings.HasPrefix(idLower, "safari_ipad_")) {
			continue
		}

		remainder := strings.TrimPrefix(idLower, prefix)
		major, ok := parseLeadingMajor(remainder)
		if !ok {
			continue
		}

		if major > latestMajor {
			latestMajor = major
			latestID = id
		}

		if hasTargetMajor && major <= targetMajor && major > bestUnderOrEqualMajor {
			bestUnderOrEqualMajor = major
			bestUnderOrEqualID = id
		}
	}

	if bestUnderOrEqualID != "" {
		return bestUnderOrEqualID
	}
	return latestID
}

func parseLeadingMajor(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	idx := 0
	for idx < len(value) {
		c := value[idx]
		if c < '0' || c > '9' {
			break
		}
		idx++
	}
	if idx == 0 {
		return 0, false
	}

	major, err := strconv.Atoi(value[:idx])
	if err != nil {
		return 0, false
	}
	return major, true
}

// getClientHelloID 获取浏览器 ClientHello ID
func getClientHelloID(browserType string) tls.ClientHelloID {
	switch browserType {
	case "chrome":
		return tls.HelloChrome_Auto
	case "firefox":
		return tls.HelloFirefox_Auto
	case "safari":
		return tls.HelloSafari_Auto
	case "edge":
		return tls.HelloChrome_Auto
	case "opera":
		return tls.HelloChrome_Auto
	default:
		return tls.HelloChrome_Auto
	}
}

func convertTLSVersion(version uint16) uint16 {
	switch version {
	case 0x0301:
		return tls.VersionTLS10
	case 0x0302:
		return tls.VersionTLS11
	case 0x0303:
		return tls.VersionTLS12
	case 0x0304:
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

func convertCipherSuites(suites []uint16) []uint16 {
	result := make([]uint16, 0, len(suites))
	for _, suite := range suites {
		switch suite {
		case 0x1301:
			result = append(result, tls.TLS_AES_128_GCM_SHA256)
		case 0x1302:
			result = append(result, tls.TLS_AES_256_GCM_SHA384)
		case 0x1303:
			result = append(result, tls.TLS_CHACHA20_POLY1305_SHA256)
		case 0xc02b:
			result = append(result, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
		case 0xc02f:
			result = append(result, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
		case 0xc02c:
			result = append(result, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384)
		case 0xc030:
			result = append(result, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		default:
			result = append(result, suite)
		}
	}
	return result
}
