// package config stores configuration around
package afconfig

import (
	"os"
	"path"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Config struct {
	AppName        string
	AppNameConfig  string
	DefaultBrowser string
	// used to override the builtin filepaths for custom installation locations
	CustomBrowserPath      string
	CustomSSOBrowserPath   string
	Keyring                *KeyringConfig `toml:",omitempty"`
	Ordering               string
	ExportCredentialSuffix string
}

type KeyringConfig struct {
	Backend                 *string `toml:",omitempty"`
	KeychainName            *string `toml:",omitempty"`
	FileDir                 *string `toml:",omitempty"`
	LibSecretCollectionName *string `toml:",omitempty"`
}

func NewLoadedConfig() (Config, error) {
	cfg := NewDefaultConfig()
	err := cfg.Load()
	return cfg, err
}

// NewDefaultConfig returns a config with OS specific defaults populated
func NewDefaultConfig() Config {
	cfg := Config{AppName: "aws-fuzzy", AppNameConfig: "aws_fuzzy"}

	// macos devices should default to the keychain backend
	if runtime.GOOS == "darwin" {
		keychain := "keychain"
		cfg.Keyring = &KeyringConfig{
			Backend: &keychain,
		}
	}
	return cfg
}

// checks and or creates the config folder on startup
func (c Config) SetupConfigFolder() error {
	configFolder, err := c.ConfigFolder()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		err := os.Mkdir(configFolder, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Config) ConfigFolder() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// check if the config folder already exists
	return path.Join(home, "."+c.AppName), nil
}

func (c *Config) Load() error {
	configFolder, err := c.ConfigFolder()
	if err != nil {
		return err
	}
	configFilePath := path.Join(configFolder, "config")

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	//c := NewDefaultConfig()

	_, err = toml.NewDecoder(file).Decode(c)
	return err
}

func (c *Config) Save() error {
	configFolder, err := c.ConfigFolder()
	if err != nil {
		return err
	}
	configFilePath := path.Join(configFolder, "config")

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	return toml.NewEncoder(file).Encode(c)
}
