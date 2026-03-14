package waf

import (
	"fmt"
	"net/http"
	"time"
)

func (w *WAF) generateChallengeToken(clientIP string) string {
	// TODO: Implement encrypted token generation
	return fmt.Sprintf("%s_%d", clientIP, time.Now().Unix())
}

func (w *WAF) serveChallenge(rw http.ResponseWriter, req *http.Request, result *WAFResult) {
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)

	if len(w.config.ChallengeHTML) > 0 {
		rw.Write(w.config.ChallengeHTML)
	} else {
		rw.Write([]byte(defaultChallengePage))
	}
}

func (w *WAF) serveBlock(rw http.ResponseWriter, req *http.Request, result *WAFResult) {
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusForbidden)

	if len(w.config.BlockResponse) > 0 {
		rw.Write(w.config.BlockResponse)
	} else {
		rw.Write([]byte(defaultBlockPage))
	}
}

const defaultChallengePage = `<!DOCTYPE html>
<html>
<head><title>Security Check</title></head>
<body>
<h1>Security Verification</h1>
<p>Please wait while we verify your request...</p>
<script>
// JS Challenge logic here
setTimeout(function() {
    document.cookie = "waf_verified=1; max-age=600";
    location.reload();
}, 3000);
</script>
</body>
</html>`

const defaultBlockPage = `<!DOCTYPE html>
<html>
<head><title>Access Denied</title></head>
<body>
<h1>Access Denied</h1>
<p>Your request has been blocked due to suspicious activity.</p>
<p>If you believe this is an error, please contact support.</p>
</body>
</html>`
