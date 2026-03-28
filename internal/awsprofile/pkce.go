package awsprofile

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	"github.com/AndreZiviani/aws-fuzzy/internal/securestorage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/common-fate/clio"
)

const (
	// authorizationCallbackTimeout is the maximum time to wait for the user to
	// complete browser-based authentication before giving up.
	authorizationCallbackTimeout = 5 * time.Minute

	// callbackPath is the OAuth callback path used for the redirect URI.
	callbackPath = "/oauth/callback"
)

// Per RFC 7636, the code verifier must be 43-128 unreserved characters.
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating code verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

type callbackResult struct {
	Code string
	Err  error
}

type callbackPageData struct {
	Error       string
	Description string
}

var callbackErrorTmpl = template.Must(template.New("error").Parse(`<!DOCTYPE html>
<html>
<head><title>Authentication Failed</title></head>
<body style="font-family: system-ui, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f8f9fa;">
<div style="text-align: center; padding: 2rem;">
<h1 style="color: #dc2626;">Authentication Failed</h1>
<p>Error: {{.Error}}</p>
<p>{{.Description}}</p>
<p>Please close this window and try again.</p>
</div>
</body>
</html>`))

const callbackSuccessHTML = `<!DOCTYPE html>
<html>
<head><title>Authentication Successful</title></head>
<body style="font-family: system-ui, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f8f9fa;">
<div style="text-align: center; padding: 2rem;">
<h1 style="color: #16a34a;">Authentication Successful</h1>
<p>You have successfully authenticated with AWS IAM Identity Center.</p>
<p>You can close this window and return to your terminal.</p>
</div>
</body>
</html>`

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
}

func writeErrorPage(w http.ResponseWriter, errCode, description string) {
	setSecurityHeaders(w)
	w.WriteHeader(http.StatusBadRequest)
	_ = callbackErrorTmpl.Execute(w, callbackPageData{
		Error:       errCode,
		Description: description,
	})
}

// newCallbackHandler creates the HTTP handler for the OAuth callback.
// It validates the state parameter against CSRF attacks, extracts the authorization code,
// and sends the result on the provided channel. Only the first request is processed (sync.Once).
func newCallbackHandler(expectedState string, resultCh chan<- callbackResult) http.Handler {
	var once sync.Once
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		// Only accept GET requests per OAuth 2.0 spec
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Only process the first callback request. Subsequent requests
		// (browser retries, favicon fetches) are ignored.
		var handled bool
		once.Do(func() {
			handled = true
			query := r.URL.Query()

			if errParam := query.Get("error"); errParam != "" {
				errDesc := query.Get("error_description")
				writeErrorPage(w, errParam, errDesc)
				resultCh <- callbackResult{Err: fmt.Errorf("%s: %s", errParam, errDesc)}
				return
			}

			state := query.Get("state")
			if state != expectedState {
				writeErrorPage(w, "state_mismatch", "The state parameter did not match. This may indicate a CSRF attack.")
				resultCh <- callbackResult{Err: errors.New("OAuth state parameter mismatch")}
				return
			}

			code := query.Get("code")
			if code == "" {
				writeErrorPage(w, "missing_code", "No authorization code was received.")
				resultCh <- callbackResult{Err: errors.New("no authorization code received")}
				return
			}

			setSecurityHeaders(w)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(callbackSuccessHTML))
			resultCh <- callbackResult{Code: code}
		})

		if !handled {
			http.Error(w, "Authorization already processed", http.StatusConflict)
		}
	})
	return mux
}

// startCallbackServer starts a temporary HTTP server on localhost to receive the PKCE callback.
// It returns the port, a channel that will receive the authorization code, and a shutdown function.
func startCallbackServer(expectedState string) (int, <-chan callbackResult, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, nil, nil, fmt.Errorf("starting callback server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	resultCh := make(chan callbackResult, 1)

	server := &http.Server{
		Handler:      newCallbackHandler(expectedState, resultCh),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			clio.Debugf("OAuth callback server error: %s", err)
		}
	}()

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}

	return port, resultCh, shutdown, nil
}

// buildAuthorizationURL constructs the PKCE authorization URL per RFC 6749 and RFC 7636.
func buildAuthorizationURL(authEndpoint, clientID, redirectURI, codeChallenge, state string, scopes []string) (string, error) {
	u, err := url.Parse(authEndpoint)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	q.Set("scope", strings.Join(scopes, " "))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// SSOPKCEFlowFromStartUrl performs the PKCE authorization code flow to retrieve an SSO token.
func SSOPKCEFlowFromStartUrl(ctx context.Context, cfg aws.Config, startUrl string, profile string) (*securestorage.SSOToken, error) {
	if cfg.Region == "" {
		return nil, errors.New("AWS region is required for authorization code flow")
	}

	ssooidcClient := ssooidc.NewFromConfig(cfg)

	// Register client with authorization_code grant type.
	// The redirect URI for registration uses the portless form. Per RFC 8252
	// Section 7.3, authorization servers MUST allow any port to be specified for
	// loopback redirect URIs. AWS IAM Identity Center implements this exemption:
	// the portless URI is registered, but the actual redirect uses a port-specific URI.
	reg, err := getOrRegisterClient(ctx, ssooidcClient, cfg, startUrl, []string{GrantTypeAuthCode, GrantTypeRefreshToken}, &ClientRegistrationOpts{
		IssuerUrl:    startUrl,
		RedirectUris: []string{"http://127.0.0.1" + callbackPath},
	})
	if err != nil {
		return nil, fmt.Errorf("registering PKCE client: %w", err)
	}

	// Determine the authorization endpoint. The RegisterClient API may return it,
	// but many regions don't include it in the response. Fall back to the standard
	// regional endpoint pattern used by the AWS CLI.
	authorizationEndpoint := fmt.Sprintf("https://oidc.%s.amazonaws.com/authorize", cfg.Region)
	if reg.AuthorizationEndpoint != "" {
		authorizationEndpoint = reg.AuthorizationEndpoint
	} else {
		clio.Debugf("authorization endpoint not returned by RegisterClient, using regional fallback: %s", authorizationEndpoint)
	}

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	state, err := generateState()
	if err != nil {
		return nil, err
	}

	port, resultCh, shutdown, err := startCallbackServer(state)
	if err != nil {
		return nil, err
	}

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d%s", port, callbackPath)
	authURL, err := buildAuthorizationURL(authorizationEndpoint, reg.ClientID, redirectURI, codeChallenge, state, []string{securestorage.ScopeAccountAccess})
	if err != nil {
		return nil, fmt.Errorf("building authorization URL: %w", err)
	}

	afcfg, err := afconfig.NewLoadedConfig()
	if err != nil {
		return nil, err
	}

	clio.Info("Opening browser for authentication...")
	if err := openBrowserForSSO(afcfg, authURL, profile, false); err != nil {
		return nil, fmt.Errorf("opening browser: %w", err)
	}

	clio.Info("Awaiting authentication in the browser...")
	clio.Info("You will be prompted to authenticate and approve access")

	var result callbackResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		shutdown()
		return nil, ctx.Err()
	case <-time.After(authorizationCallbackTimeout):
		shutdown()
		return nil, errors.New("timed out waiting for authorization callback")
	}

	// Free the port before the token exchange network round-trip.
	shutdown()

	if result.Err != nil {
		return nil, fmt.Errorf("authorization failed: %w", result.Err)
	}

	token, err := ssooidcClient.CreateToken(ctx, &ssooidc.CreateTokenInput{
		ClientId:     &reg.ClientID,
		ClientSecret: &reg.ClientSecret,
		GrantType:    aws.String(GrantTypeAuthCode),
		Code:         &result.Code,
		CodeVerifier: &codeVerifier,
		RedirectUri:  &redirectURI,
	})
	if err != nil {
		return nil, fmt.Errorf("exchanging authorization code: %w", err)
	}

	return newSSOTokenFromResponse(token, reg, cfg.Region), nil
}
