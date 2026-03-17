// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keyring

import (
	"encoding/base64"
	"encoding/hex"
	"os/exec"
	"strings"
)

const (
	execPathKeychain = "/usr/bin/security"

	// encodingPrefix is a well-known prefix added to strings encoded by Set.
	encodingPrefix       = "go-keyring-encoded:"
	base64EncodingPrefix = "go-keyring-base64:"
)

type macOSXKeychain struct{}

// Get returns a secret for username under service.
func (k macOSXKeychain) Get(service, username string) (string, error) {
	out, err := exec.Command(
		execPathKeychain,
		"find-generic-password",
		"-s", service,
		"-wa", username).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "could not be found") {
			err = ErrKeyringNotFound
		}
		return "", err
	}

	trimStr := strings.TrimSpace(string(out[:]))
	// if the string has the well-known prefix, assume it's encoded
	if strings.HasPrefix(trimStr, encodingPrefix) {
		dec, err := hex.DecodeString(trimStr[len(encodingPrefix):])
		return string(dec), err
	} else if strings.HasPrefix(trimStr, base64EncodingPrefix) {
		dec, err := base64.StdEncoding.DecodeString(trimStr[len(base64EncodingPrefix):])
		return string(dec), err
	}

	return trimStr, nil
}

// Set stores a secret for username under service.
func (k macOSXKeychain) Set(service, username, password string) error {
	// if the added secret has multiple lines or some non ascii,
	// osx will hex encode it on return. To avoid getting garbage, we
	// encode all passwords
	password = base64EncodingPrefix + base64.StdEncoding.EncodeToString([]byte(password))

	// Delete any existing item before creating a new one. Using -U (update) on
	// an item created with different ACLs triggers a macOS keychain password
	// prompt to "change access permissions". Creating fresh with -A avoids this.
	_ = exec.Command(execPathKeychain, "delete-generic-password", "-s", service, "-a", username).Run()

	out, err := exec.Command(
		execPathKeychain,
		"add-generic-password",
		"-A",
		"-s", service,
		"-a", username,
		"-w", password,
	).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "Data too long") || strings.Contains(string(out), "value is too large") {
			return ErrSetDataTooBig
		}
		return err
	}

	return nil
}

// Delete removes a secret for username under service.
func (k macOSXKeychain) Delete(service, username string) error {
	out, err := exec.Command(
		execPathKeychain,
		"delete-generic-password",
		"-s", service,
		"-a", username).CombinedOutput()
	if strings.Contains(string(out), "could not be found") {
		err = ErrKeyringNotFound
	}
	return err
}

// DeleteAll removes all secrets under service.
func (k macOSXKeychain) DeleteAll(service string) error {
	// if service is empty, do nothing otherwise it might accidentally delete all secrets
	if service == "" {
		return ErrKeyringNotFound
	}
	// Delete each secret in a while loop until there is no more left
	// under the service
	for {
		out, err := exec.Command(
			execPathKeychain,
			"delete-generic-password",
			"-s", service).CombinedOutput()
		if strings.Contains(string(out), "could not be found") {
			return nil
		} else if err != nil {
			return err
		}
	}

}

func init() {
	provider = macOSXKeychain{}
}
