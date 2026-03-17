package keyring

import (
	"strings"
	"syscall"

	"github.com/danieljoos/wincred"
)

type windowsKeychain struct{}

// Get returns a secret for username under service.
func (k windowsKeychain) Get(service, username string) (string, error) {
	cred, err := wincred.GetGenericCredential(k.credName(service, username))
	if err != nil {
		if err == syscall.ERROR_NOT_FOUND {
			return "", ErrKeyringNotFound
		}
		return "", err
	}

	return string(cred.CredentialBlob), nil
}

// Set stores password for username under service.
func (k windowsKeychain) Set(service, username, password string) error {
	// password may not exceed 2560 bytes (https://github.com/jaraco/keyring/issues/540#issuecomment-968329967)
	if len(password) > 2560 {
		return ErrSetDataTooBig
	}

	// service may not exceed 512 bytes (might need more testing)
	if len(service) >= 512 {
		return ErrSetDataTooBig
	}

	// service may not exceed 32k but problems occur before that
	// so we limit it to 30k
	if len(service) > 1024*30 {
		return ErrSetDataTooBig
	}

	cred := wincred.NewGenericCredential(k.credName(service, username))
	cred.UserName = username
	cred.CredentialBlob = []byte(password)
	return cred.Write()
}

// Delete removes a secret for username under service.
func (k windowsKeychain) Delete(service, username string) error {
	cred, err := wincred.GetGenericCredential(k.credName(service, username))
	if err != nil {
		if err == syscall.ERROR_NOT_FOUND {
			return ErrKeyringNotFound
		}
		return err
	}

	return cred.Delete()
}

// DeleteAll removes all secrets under service.
func (k windowsKeychain) DeleteAll(service string) error {
	// if service is empty, do nothing otherwise it might accidentally delete all secrets
	if service == "" {
		return ErrKeyringNotFound
	}

	creds, err := wincred.List()
	if err != nil {
		return err
	}

	prefix := k.credName(service, "")

	for _, cred := range creds {
		if strings.HasPrefix(cred.TargetName, prefix) {
			genericCred, err := wincred.GetGenericCredential(cred.TargetName)
			if err != nil {
				if err != syscall.ERROR_NOT_FOUND {
					return err
				}
			} else {
				err := genericCred.Delete()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// credName combines service and username to a single string.
func (k windowsKeychain) credName(service, username string) string {
	return service + ":" + username
}

func init() {
	provider = windowsKeychain{}
}
