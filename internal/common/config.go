package common

import (
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/common-fate/granted/pkg/config"
)

var (
	ConfigDir = fmt.Sprintf("%s/.aws-fuzzy/", UserHomeDir)
)

func ConfigLoad() (*config.Config, error) {
	dir := ConfigDir

	configFile, err := os.OpenFile(path.Join(dir, "config"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	c := config.NewDefaultConfig()

	_, err = toml.NewDecoder(configFile).Decode(&c)
	if err != nil {
		// if there is an error just reset the file
		return &c, nil
	}
	return &c, nil
}
