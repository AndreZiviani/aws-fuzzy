//go:build (dragonfly && cgo) || (freebsd && cgo) || linux || netbsd || openbsd

package keyring

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
	ss "github.com/AndreZiviani/aws-fuzzy/internal/keyring/secret_service"
)

type secretServiceProvider struct{}

// Set stores pass for user under service.
func (s secretServiceProvider) Set(service, user, pass string) error {
	svc, err := ss.NewSecretService()
	if err != nil {
		return err
	}

	// open a session
	session, err := svc.OpenSession()
	if err != nil {
		return err
	}
	defer svc.Close(session)

	attributes := map[string]string{
		"username": user,
		"service":  service,
	}

	secret := ss.NewSecret(session.Path(), pass)

	collection := svc.GetLoginCollection()

	err = svc.Unlock(collection.Path())
	if err != nil {
		return err
	}

	err = svc.CreateItem(collection,
		fmt.Sprintf("Password for '%s' on '%s'", user, service),
		attributes, secret)
	if err != nil {
		return err
	}

	return nil
}

// findItem looks up an item by service and user.
func (s secretServiceProvider) findItem(svc *ss.SecretService, service, user string) (dbus.ObjectPath, error) {
	collection := svc.GetLoginCollection()

	search := map[string]string{
		"username": user,
		"service":  service,
	}

	err := svc.Unlock(collection.Path())
	if err != nil {
		return "", err
	}

	results, err := svc.SearchItems(collection, search)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", ErrKeyringNotFound
	}

	return results[0], nil
}

// findServiceItems looks up all items by service.
func (s secretServiceProvider) findServiceItems(svc *ss.SecretService, service string) ([]dbus.ObjectPath, error) {
	collection := svc.GetLoginCollection()

	search := map[string]string{
		"service": service,
	}

	err := svc.Unlock(collection.Path())
	if err != nil {
		return []dbus.ObjectPath{}, err
	}

	results, err := svc.SearchItems(collection, search)
	if err != nil {
		return []dbus.ObjectPath{}, err
	}

	if len(results) == 0 {
		return []dbus.ObjectPath{}, ErrKeyringNotFound
	}

	return results, nil
}

// Get returns a secret for user under service.
func (s secretServiceProvider) Get(service, user string) (string, error) {
	svc, err := ss.NewSecretService()
	if err != nil {
		return "", err
	}

	item, err := s.findItem(svc, service, user)
	if err != nil {
		return "", err
	}

	// open a session
	session, err := svc.OpenSession()
	if err != nil {
		return "", err
	}
	defer svc.Close(session)

	// unlock if individual item is locked
	err = svc.Unlock(item)
	if err != nil {
		return "", err
	}

	secret, err := svc.GetSecret(item, session.Path())
	if err != nil {
		return "", err
	}

	return string(secret.Value), nil
}

// Delete removes a secret for user under service.
func (s secretServiceProvider) Delete(service, user string) error {
	svc, err := ss.NewSecretService()
	if err != nil {
		return err
	}

	item, err := s.findItem(svc, service, user)
	if err != nil {
		return err
	}

	return svc.Delete(item)
}

// DeleteAll removes all secrets under service.
func (s secretServiceProvider) DeleteAll(service string) error {
	// if service is empty, do nothing otherwise it might accidentally delete all secrets
	if service == "" {
		return ErrKeyringNotFound
	}

	svc, err := ss.NewSecretService()
	if err != nil {
		return err
	}
	// find all items for the service
	items, err := s.findServiceItems(svc, service)
	if err != nil {
		if err == ErrKeyringNotFound {
			return nil
		}
		return err
	}
	for _, item := range items {
		err = svc.Delete(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	provider = secretServiceProvider{}
}
