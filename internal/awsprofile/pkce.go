package awsprofile

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	"github.com/AndreZiviani/aws-fuzzy/internal/securestorage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/common-fate/clio"
)

// generateCodeVerifier creates a cryptographically random PKCE code verifier.
// Per RFC 7636, it must be 43-128 characters of [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~".
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating code verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge computes the S256 code challenge from the verifier.
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// callbackResult holds the result of the PKCE callback.
type callbackResult struct {
	Code string
	Err  error
}

// startCallbackServer starts a temporary HTTP server on localhost to receive the PKCE callback.
// It returns the port, a channel that will receive the authorization code, and a shutdown function.
func startCallbackServer(ctx context.Context) (int, <-chan callbackResult, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, nil, nil, fmt.Errorf("starting callback server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errParam := r.URL.Query().Get("error")

		if errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><body><h1>Authentication Failed</h1><p>%s: %s</p><p>You can close this tab.</p></body></html>", errParam, errDesc)
			resultCh <- callbackResult{Err: fmt.Errorf("authorization error: %s - %s", errParam, errDesc)}
			return
		}

		if code == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, "<html><body><h1>Error</h1><p>No authorization code received.</p></body></html>")
			resultCh <- callbackResult{Err: fmt.Errorf("no authorization code in callback")}
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Authentication Successful</h1><p>You can close this tab and return to the terminal.</p></body></html>")
		resultCh <- callbackResult{Code: code}
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			resultCh <- callbackResult{Err: fmt.Errorf("callback server error: %w", err)}
		}
	}()

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}

	return port, resultCh, shutdown, nil
}

// buildAuthorizationURL constructs the PKCE authorization URL.
func buildAuthorizationURL(authEndpoint, clientID, redirectURI, codeChallenge string) string {
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"scopes":                {securestorage.ScopeAccountAccess},
	}
	return authEndpoint + "?" + params.Encode()
}

// SSOPKCEFlowFromStartUrl performs the PKCE authorization code flow to retrieve an SSO token.
func SSOPKCEFlowFromStartUrl(ctx context.Context, cfg aws.Config, startUrl string, profile string, printOnly bool) (*securestorage.SSOToken, error) {
	if printOnly {
		// PKCE requires a local server to receive the callback; fall back to device code
		return nil, fmt.Errorf("PKCE flow not supported in print-only mode")
	}

	ssooidcClient := ssooidc.NewFromConfig(cfg)

	reg, err := getOrRegisterClient(ctx, ssooidcClient, cfg, startUrl, []string{GrantTypeAuthCode, GrantTypeRefreshToken})
	if err != nil {
		return nil, fmt.Errorf("registering PKCE client: %w", err)
	}

	if reg.AuthorizationEndpoint == "" {
		return nil, fmt.Errorf("no authorization endpoint returned from client registration")
	}

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	port, resultCh, shutdown, err := startCallbackServer(ctx)
	if err != nil {
		return nil, err
	}
	defer shutdown()

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d", port)
	authURL := buildAuthorizationURL(reg.AuthorizationEndpoint, reg.ClientID, redirectURI, codeChallenge)

	afcfg, err := afconfig.NewLoadedConfig()
	if err != nil {
		return nil, err
	}

	clio.Info("Opening browser for authentication...")
	if err := openBrowserForSSO(afcfg, authURL, profile, false); err != nil {
		return nil, fmt.Errorf("opening browser: %w", err)
	}

	clio.Info("Awaiting authentication in the browser...")

	// Wait for the callback with a timeout
	select {
	case result := <-resultCh:
		if result.Err != nil {
			return nil, result.Err
		}

		// Exchange the authorization code for tokens
		token, err := ssooidcClient.CreateToken(ctx, &ssooidc.CreateTokenInput{
			ClientId:     &reg.ClientID,
			ClientSecret: &reg.ClientSecret,
			GrantType:    aws.String(GrantTypeAuthCode),
			Code:         &result.Code,
			RedirectUri:  &redirectURI,
			CodeVerifier: &codeVerifier,
			Scope:        []string{securestorage.ScopeAccountAccess},
		})
		if err != nil {
			return nil, fmt.Errorf("exchanging authorization code: %w", err)
		}

		return &securestorage.SSOToken{
			AccessToken:           aws.ToString(token.AccessToken),
			Expiry:                time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
			ClientID:              reg.ClientID,
			ClientSecret:          reg.ClientSecret,
			RegistrationExpiresAt: reg.RegistrationExpiresAt,
			Region:                cfg.Region,
			RefreshToken:          token.RefreshToken,
		}, nil

	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("timed out waiting for browser authentication")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
