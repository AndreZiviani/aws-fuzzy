package securestorage

import (
	"encoding/json"
	"errors"

	"github.com/AndreZiviani/aws-fuzzy/internal/keyring"
)

type SecureStorage struct {
	StoragePrefix string
	StorageSuffix string
	Debug         bool
}

// serviceName returns the keyring service name for this storage instance.
func (s *SecureStorage) serviceName() string {
	return s.StoragePrefix + s.StorageSuffix
}

func (s *SecureStorage) Retrieve(key string, target any) error {
	val, err := keyring.Get(s.serviceName(), key)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyringNotFound) {
			return ErrNotFound
		}
		return err
	}
	return json.Unmarshal([]byte(val), target)
}

func (s *SecureStorage) Store(key string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return keyring.Set(s.serviceName(), key, string(b))
}

func (s *SecureStorage) Clear(key string) error {
	err := keyring.Delete(s.serviceName(), key)
	if errors.Is(err, keyring.ErrKeyringNotFound) {
		return nil
	}
	return err
}

var ErrNotFound = errors.New("key not found in secure storage")
