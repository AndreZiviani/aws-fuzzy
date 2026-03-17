package keyring

import (
	"errors"
	"time"
)

var (
	// ErrKeyringNotFound is the expected error if the secret isn't found in the
	// keyring.
	ErrKeyringNotFound = errors.New("secret not found in keyring")
	// ErrSetDataTooBig is returned if `Set` was called with too much data.
	// On MacOS: The combination of service, username & password should not exceed ~3000 bytes
	// On Windows: The service is limited to 32KiB while the password is limited to 2560 bytes
	// On Linux/Unix: There is no theoretical limit but performance suffers with big values (>100KiB)
	ErrSetDataTooBig  = errors.New("data passed to Set was too big")
	ErrKeyringTimeout = errors.New("timeout while accessing keyring")

	// provider is set in platform-specific init functions (e.g. keyring_unix.go)
	// and receives the service value from this package internals.
	provider Keyring = fallbackServiceProvider{}
)

// Keyring defines the internal provider contract used by OS-specific backends.
//
// The service argument is provider-level only. Public functions in this package
// always pass the caller-supplied service name.
type Keyring interface {
	// Set stores secret for user under service.
	Set(service, user, secret string) error
	// Get returns secret for user under service.
	Get(service, user string) (string, error)
	// Delete removes secret for user under service.
	Delete(service, user string) error
	// DeleteAll removes all secrets under service.
	DeleteAll(service string) error
}

// Set secret in keyring for user under the given service.
func Set(service, user, secret string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- provider.Set(service, user, secret)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return ErrKeyringTimeout
	}
}

// Get secret from keyring for user under the given service.
func Get(service, user string) (string, error) {
	ch := make(chan struct {
		val string
		err error
	}, 1)
	go func() {
		defer close(ch)
		val, err := provider.Get(service, user)
		ch <- struct {
			val string
			err error
		}{val, err}
	}()
	select {
	case res := <-ch:
		return res.val, res.err
	case <-time.After(3 * time.Second):
		return "", ErrKeyringTimeout
	}
}

// Delete secret from keyring for user under the given service.
func Delete(service, user string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- provider.Delete(service, user)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return ErrKeyringTimeout
	}
}

// DeleteAll deletes all secrets for the given service.
func DeleteAll(service string) error {
	return provider.DeleteAll(service)
}
