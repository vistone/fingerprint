package client

import (
	"context"

	fhttp "github.com/bogdanfinn/fhttp"
)

// RoundTrip executes request and returns a unified *fhttp.Response.
func (st *SmartTransport) RoundTrip(req *fhttp.Request) (*fhttp.Response, error) {
	ctx := req.Context()
	host := st.requestHost(req)

	resp, handled, err := st.tryCachedHTTP1(ctx, req, host)
	if handled {
		return resp, err
	}

	return st.tryHTTP2ThenFallback(ctx, req, host)
}

func (st *SmartTransport) requestHost(req *fhttp.Request) string {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	return host
}

func (st *SmartTransport) cachedProtocol(host string) string {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.hostProtocolCache[host]
}

func (st *SmartTransport) setCachedProtocol(host, protocol string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.hostProtocolCache[host] = protocol
}

func (st *SmartTransport) tryCachedHTTP1(ctx context.Context, req *fhttp.Request, host string) (*fhttp.Response, bool, error) {
	if st.cachedProtocol(host) != "http/1.1" {
		return nil, false, nil
	}

	resp, err := st.roundTripHTTP1(ctx, req)
	if err == nil && !shouldRetryWithHTTP1(resp) {
		return resp, true, nil
	}
	if st.strictFingerprint {
		if err != nil {
			return nil, true, err
		}
		return resp, true, nil
	}

	resp, err = st.roundTripHTTP1Compat(ctx, req)
	return resp, true, err
}

func (st *SmartTransport) tryHTTP2ThenFallback(ctx context.Context, req *fhttp.Request, host string) (*fhttp.Response, error) {
	resp, err := st.roundTripHTTP2(ctx, req)
	if err == nil {
		return st.handleSuccessfulHTTP2(ctx, req, host, resp)
	}

	errType := classifyError(err)
	if errType == ErrorTypeTimeout || errType == ErrorTypeCanceled {
		return nil, err
	}
	return st.handleFailedHTTP2(ctx, req, host, err)
}

func (st *SmartTransport) handleSuccessfulHTTP2(ctx context.Context, req *fhttp.Request, host string, resp *fhttp.Response) (*fhttp.Response, error) {
	if shouldRetryWithHTTP1(resp) {
		return st.retryFromHTTP2Response(ctx, req, host, resp)
	}

	st.setCachedProtocol(host, "h2")
	return resp, nil
}

func (st *SmartTransport) retryFromHTTP2Response(ctx context.Context, req *fhttp.Request, host string, h2Resp *fhttp.Response) (*fhttp.Response, error) {
	h1Resp, h1Err := st.roundTripHTTP1(ctx, req)
	if h1Err == nil && !shouldRetryWithHTTP1(h1Resp) {
		if h2Resp.Body != nil {
			_ = h2Resp.Body.Close()
		}
		st.setCachedProtocol(host, "http/1.1")
		return h1Resp, nil
	}

	if st.strictFingerprint {
		st.setCachedProtocol(host, "h2")
		return h2Resp, nil
	}

	compatResp, compatErr := st.roundTripHTTP1Compat(ctx, req)
	if compatErr == nil {
		if h2Resp.Body != nil {
			_ = h2Resp.Body.Close()
		}
		if h1Resp != nil && h1Resp.Body != nil {
			_ = h1Resp.Body.Close()
		}
		st.setCachedProtocol(host, "http/1.1")
		return compatResp, nil
	}

	st.setCachedProtocol(host, "h2")
	return h2Resp, nil
}

func (st *SmartTransport) handleFailedHTTP2(ctx context.Context, req *fhttp.Request, host string, h2Err error) (*fhttp.Response, error) {
	resp, err := st.roundTripHTTP1(ctx, req)
	if err == nil {
		st.setCachedProtocol(host, "http/1.1")
		return resp, nil
	}

	if !st.strictFingerprint && shouldFallbackToHTTP1Compat(err) {
		compatResp, compatErr := st.roundTripHTTP1Compat(ctx, req)
		if compatErr == nil {
			st.setCachedProtocol(host, "http/1.1")
			return compatResp, nil
		}
	}

	// if both protocols fail, return first error (HTTP/2 error)
	return nil, h2Err
}
