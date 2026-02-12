package securestorage

import (
	"time"

	"github.com/common-fate/clio"
	"github.com/pkg/errors"
)

type EKSTokenSecureStorage struct {
	SecureStorage SecureStorage
}

func NewSecureEKSTokenStorage() EKSTokenSecureStorage {
	return EKSTokenSecureStorage{
		SecureStorage: SecureStorage{
			StoragePrefix: "aws-fuzzy",
			StorageSuffix: "-eks-tokens",
			Debug:         false,
		},
	}
}

type EKSToken struct {
	Token      string    `json:"token"`
	Expiration time.Time `json:"expiration"`
}

// GetValidEKSToken retrieves a cached EKS token if it exists and is not expired.
// Returns nil if the token is not found or has expired.
func (s *EKSTokenSecureStorage) GetValidEKSToken(cacheKey string) (*EKSToken, error) {
	var t EKSToken
	err := s.SecureStorage.Retrieve(cacheKey, &t)
	if err != nil {
		return nil, err
	}

	// Check expiration with a small buffer to avoid edge cases
	if t.Expiration.Before(time.Now().Add(30 * time.Second)) {
		s.ClearEKSToken(cacheKey)
		return nil, nil
	}

	return &t, nil
}

// StoreEKSToken caches an EKS token in secure storage.
func (s *EKSTokenSecureStorage) StoreEKSToken(cacheKey string, token EKSToken) {
	err := s.SecureStorage.Store(cacheKey, token)
	if err != nil {
		clio.Debugf("%s\n", errors.Wrap(err, "writing EKS token to credentials cache").Error())
	}
}

// ClearEKSToken removes a cached EKS token from secure storage.
func (s *EKSTokenSecureStorage) ClearEKSToken(cacheKey string) {
	err := s.SecureStorage.Clear(cacheKey)
	if err != nil {
		clio.Debugf("%s\n", errors.Wrap(err, "clearing EKS token from the credentials cache").Error())
	}
}
