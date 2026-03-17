package securestorage

import (
	"time"

	"github.com/common-fate/clio"
)

// registrationExpiryBuffer is how early before expiry we consider a registration invalid,
// ensuring we don't use a registration that's about to expire.
const registrationExpiryBuffer = 24 * time.Hour

type ClientRegistration struct {
	ClientID              string    `json:"clientId"`
	ClientSecret          string    `json:"clientSecret"`
	RegistrationExpiresAt time.Time `json:"registrationExpiresAt"`
	Region                string    `json:"region"`
	// AuthorizationEndpoint is returned by RegisterClient and used for PKCE flows.
	AuthorizationEndpoint string `json:"authorizationEndpoint,omitempty"`
	// TokenEndpoint is returned by RegisterClient and used for PKCE flows.
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`
}

type ClientRegistrationSecureStorage struct {
	SecureStorage SecureStorage
}

func NewSecureClientRegistrationStorage() ClientRegistrationSecureStorage {
	return ClientRegistrationSecureStorage{
		SecureStorage: SecureStorage{
			StoragePrefix: "aws-fuzzy",
			StorageSuffix: "-sso-registrations",
			Debug:         false,
		},
	}
}

// GetValidRegistration retrieves a cached client registration if it exists and is not expired.
func (s *ClientRegistrationSecureStorage) GetValidRegistration(key string) *ClientRegistration {
	var r ClientRegistration
	err := s.SecureStorage.Retrieve(key, &r)
	if err != nil {
		return nil
	}

	if r.RegistrationExpiresAt.Before(time.Now().Add(registrationExpiryBuffer)) {
		s.ClearRegistration(key)
		return nil
	}

	return &r
}

// StoreRegistration caches a client registration in secure storage.
func (s *ClientRegistrationSecureStorage) StoreRegistration(key string, reg ClientRegistration) {
	err := s.SecureStorage.Store(key, reg)
	if err != nil {
		clio.Debugf("writing SSO client registration to cache: %s", err.Error())
	}
}

// ClearRegistration removes a cached client registration from secure storage.
func (s *ClientRegistrationSecureStorage) ClearRegistration(key string) {
	err := s.SecureStorage.Clear(key)
	if err != nil {
		clio.Debugf("clearing SSO client registration from cache: %s", err.Error())
	}
}
