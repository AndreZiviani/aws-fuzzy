package securestorage

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	ssooidctypes "github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
	"github.com/common-fate/clio"
)

type SSOTokensSecureStorage struct {
	SecureStorage SecureStorage
}

func NewSecureSSOTokenStorage() SSOTokensSecureStorage {
	return SSOTokensSecureStorage{
		SecureStorage: SecureStorage{
			StoragePrefix: "aws-fuzzy",
			StorageSuffix: "-sso-tokens",
			Debug:         false,
		},
	}
}

type SSOToken struct {
	AccessToken           string    `json:"accessToken"`
	Expiry                time.Time `json:"expiry"`
	ClientID              string    `json:"clientId,omitempty"`
	ClientSecret          string    `json:"clientSecret,omitempty"`
	RegistrationExpiresAt time.Time `json:"registrationExpiresAt,omitempty"`
	Region                string    `json:"region,omitempty"`
	RefreshToken          *string   `json:"refreshToken,omitempty"`
}

const (
	// tokenRefreshBuffer is how early before expiry we attempt to refresh the token,
	// avoiding failures at the expiry boundary.
	tokenRefreshBuffer = 5 * time.Minute

	// ScopeAccountAccess is the OAuth scope required for SSO account access and refresh tokens.
	ScopeAccountAccess = "sso:account:access"
)

// GetValidSSOToken loads and potentially refreshes an AWS SSO access token from secure storage.
// It returns nil if no token was found, or if it is expired and cannot be refreshed.
func (s *SSOTokensSecureStorage) GetValidSSOToken(ctx context.Context, profileKey string) *SSOToken {
	var t SSOToken
	err := s.SecureStorage.Retrieve(profileKey, &t)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			clio.Warnf("error retrieving IAM Identity Center token from secure storage: %s", err.Error())
		}
		return nil
	}
	now := time.Now()
	needsRefresh := t.Expiry.Before(now.Add(tokenRefreshBuffer))
	isExpired := t.Expiry.Before(now)

	if !needsRefresh {
		// token is valid and not close to expiry
		return &t
	}

	if t.RefreshToken == nil || *t.RefreshToken == "" {
		if !isExpired {
			// token is still valid but we can't refresh it; return it as-is
			return &t
		}
		return nil
	}

	if !t.RegistrationExpiresAt.IsZero() && t.RegistrationExpiresAt.Before(now) {
		clio.Warnf("SSO client registration has expired (after ~90 days). A full re-registration and device authorization will be required.")
		if !isExpired {
			return &t
		}
		return nil
	}

	// attempt to refresh the token
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		clio.Errorf("error loading default AWS config for token refresh: %s", err.Error())
		if !isExpired {
			return &t
		}
		return nil
	}

	if t.Region == "" {
		clio.Errorf("existing token had no SSO region set")
		if !isExpired {
			return &t
		}
		return nil
	}

	cfg.Region = t.Region

	client := ssooidc.NewFromConfig(cfg)

	res, err := client.CreateToken(ctx, &ssooidc.CreateTokenInput{
		ClientId:     &t.ClientID,
		ClientSecret: &t.ClientSecret,
		GrantType:    aws.String("refresh_token"),
		RefreshToken: t.RefreshToken,
		Scope:        []string{ScopeAccountAccess},
	})
	if err != nil {
		var invalidGrant *ssooidctypes.InvalidGrantException
		if errors.As(err, &invalidGrant) {
			clio.Warnf("Your IAM Identity Center portal session has expired. Re-authentication required.")
			clio.Warnf("To reduce login frequency, ask your AWS admin to increase the portal session duration in IAM Identity Center > Settings > Authentication (max 90 days).")
		} else {
			clio.Errorf("error refreshing AWS IAM Identity Center token: %s", err.Error())
		}
		if !isExpired {
			return &t
		}
		return nil
	}

	clio.Debugf("successfully refreshed IAM Identity Center access token")

	newToken := SSOToken{
		AccessToken:           *res.AccessToken,
		Expiry:                time.Now().Add(time.Duration(res.ExpiresIn) * time.Second),
		ClientID:              t.ClientID,
		ClientSecret:          t.ClientSecret,
		RegistrationExpiresAt: t.RegistrationExpiresAt,
		RefreshToken:          res.RefreshToken,
		Region:                t.Region,
	}

	// save the refreshed token to secure storage
	s.StoreSSOToken(profileKey, newToken)

	return &newToken
}

// Attempts to store the token, any errors will be logged to debug logging
func (s *SSOTokensSecureStorage) StoreSSOToken(profileKey string, ssoTokenValue SSOToken) {
	err := s.SecureStorage.Store(profileKey, ssoTokenValue)
	if err != nil {
		clio.Debugf("writing sso token to credentials cache: %s", err.Error())
	}

}

// Attempts to clear the token, any errors will be logged to debug logging
func (s *SSOTokensSecureStorage) ClearSSOToken(profileKey string) {
	err := s.SecureStorage.Clear(profileKey)
	if err != nil {
		clio.Debugf("clearing sso token from the credentials cache: %s", err.Error())
	}
}
