package awsprofile

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := generateCodeVerifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// RFC 7636: code verifier must be 43-128 characters
	if len(verifier) < 43 || len(verifier) > 128 {
		t.Errorf("verifier length %d not in range [43, 128]", len(verifier))
	}

	// Should be base64url encoded (no padding, no + or /)
	for _, c := range verifier {
		if c == '=' || c == '+' || c == '/' {
			t.Errorf("verifier contains invalid base64url character: %c", c)
		}
	}

	// Two verifiers should be different
	verifier2, err := generateCodeVerifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verifier == verifier2 {
		t.Error("two generated verifiers should not be equal")
	}
}

func TestComputeCodeChallenge_RFC7636(t *testing.T) {
	// RFC 7636 Appendix B known test vector
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := generateCodeChallenge(verifier)
	expected := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	if challenge != expected {
		t.Errorf("challenge = %q, want %q", challenge, expected)
	}

	if strings.Contains(challenge, "=") {
		t.Error("challenge contains padding character")
	}
}

func TestGenerateState(t *testing.T) {
	state, err := generateState()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 16 bytes -> 32 hex characters
	if len(state) != 32 {
		t.Errorf("state length = %d, want 32", len(state))
	}

	state2, err := generateState()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state == state2 {
		t.Error("two generated states should not be equal")
	}
}

func TestBuildAuthorizationURL(t *testing.T) {
	u, err := buildAuthorizationURL(
		"https://oidc.us-west-2.amazonaws.com/authorize",
		"client-123",
		"http://127.0.0.1:12345/oauth/callback",
		"challenge-value",
		"state-uuid",
		[]string{"sso:account:access"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain := []string{
		"response_type=code",
		"client_id=client-123",
		"state=state-uuid",
		"code_challenge=challenge-value",
		"code_challenge_method=S256",
		"scope=sso%3Aaccount%3Aaccess",
		"redirect_uri=",
	}
	for _, s := range mustContain {
		if !strings.Contains(u, s) {
			t.Errorf("URL missing %q:\n  %s", s, u)
		}
	}

	// Must be "scope" (singular per RFC 6749), not "scopes"
	if strings.Contains(u, "scopes=") {
		t.Error("URL contains 'scopes=' (plural) instead of 'scope='")
	}
}

func TestCallbackHandler_Success(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := newCallbackHandler("expected-state", result)

	req := httptest.NewRequest("GET", "/oauth/callback?code=auth-code-123&state=expected-state", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Authentication Successful") {
		t.Error("response body missing success message")
	}

	assertHeader(t, w, "Cache-Control", "no-store")
	assertHeader(t, w, "Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
	assertHeader(t, w, "X-Content-Type-Options", "nosniff")
	assertHeader(t, w, "X-Frame-Options", "DENY")

	r := <-result
	if r.Err != nil {
		t.Errorf("unexpected error: %v", r.Err)
	}
	if r.Code != "auth-code-123" {
		t.Errorf("code = %q, want %q", r.Code, "auth-code-123")
	}
}

func TestCallbackHandler_StateMismatch(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := newCallbackHandler("expected-state", result)

	req := httptest.NewRequest("GET", "/oauth/callback?code=auth-code-123&state=wrong-state", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	r := <-result
	if r.Err == nil {
		t.Fatal("expected error for state mismatch")
	}
	if !strings.Contains(r.Err.Error(), "state parameter mismatch") {
		t.Errorf("error = %q, want state mismatch", r.Err.Error())
	}
}

func TestCallbackHandler_OAuthError(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := newCallbackHandler("expected-state", result)

	req := httptest.NewRequest("GET", "/oauth/callback?error=access_denied&error_description=User+denied+access", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	r := <-result
	if r.Err == nil {
		t.Fatal("expected error for OAuth error response")
	}
	if !strings.Contains(r.Err.Error(), "access_denied") {
		t.Errorf("error = %q, want access_denied", r.Err.Error())
	}
}

func TestCallbackHandler_MissingCode(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := newCallbackHandler("expected-state", result)

	req := httptest.NewRequest("GET", "/oauth/callback?state=expected-state", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	r := <-result
	if r.Err == nil {
		t.Fatal("expected error for missing code")
	}
	if !strings.Contains(r.Err.Error(), "no authorization code") {
		t.Errorf("error = %q, want 'no authorization code'", r.Err.Error())
	}
}

func TestCallbackHandler_XSSPrevention(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := newCallbackHandler("expected-state", result)

	req := httptest.NewRequest("GET", `/oauth/callback?error=<script>alert(1)</script>&error_description=<img+onerror=alert(1)+src=x>`, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if strings.Contains(body, "<script>") {
		t.Error("response contains unescaped <script> tag")
	}
	if strings.Contains(body, "<img ") {
		t.Error("response contains unescaped <img> tag")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Error("response should contain HTML-escaped script tag")
	}
}

func TestCallbackHandler_DuplicateRequest(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := newCallbackHandler("expected-state", result)

	// First request succeeds
	req1 := httptest.NewRequest("GET", "/oauth/callback?code=auth-code-123&state=expected-state", nil)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("first request: status = %d, want %d", w1.Code, http.StatusOK)
	}
	r := <-result
	if r.Err != nil {
		t.Errorf("unexpected error: %v", r.Err)
	}

	// Second request gets 409 Conflict
	req2 := httptest.NewRequest("GET", "/oauth/callback?code=another-code&state=expected-state", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("second request: status = %d, want %d", w2.Code, http.StatusConflict)
	}
}

func TestCallbackHandler_MethodNotAllowed(t *testing.T) {
	result := make(chan callbackResult, 1)
	handler := newCallbackHandler("expected-state", result)

	req := httptest.NewRequest("POST", "/oauth/callback?code=auth-code-123&state=expected-state", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestIsHeadlessEnvironment(t *testing.T) {
	// Clear all headless indicators for a clean baseline.
	for _, env := range []string{"SSH_CLIENT", "SSH_TTY", "SSH_CONNECTION", "CI", "CODESPACES", "CLOUD_SHELL"} {
		t.Setenv(env, "")
	}

	t.Run("not headless by default", func(t *testing.T) {
		if IsHeadlessEnvironment() {
			t.Error("expected non-headless in clean environment")
		}
	})

	t.Run("SSH_CLIENT triggers headless", func(t *testing.T) {
		t.Setenv("SSH_CLIENT", "1.2.3.4 5678 22")
		if !IsHeadlessEnvironment() {
			t.Error("expected headless when SSH_CLIENT is set")
		}
	})

	t.Run("CI triggers headless", func(t *testing.T) {
		t.Setenv("CI", "true")
		if !IsHeadlessEnvironment() {
			t.Error("expected headless when CI is set")
		}
	})
}

func assertHeader(t *testing.T, w *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	got := w.Header().Get(key)
	if got != want {
		t.Errorf("header %s = %q, want %q", key, got, want)
	}
}
