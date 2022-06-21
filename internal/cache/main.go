package cache

import (
	"fmt"
	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/faabiosr/cachego"
	"github.com/faabiosr/cachego/file"
	"os"
)

var (
	CacheDir = fmt.Sprintf("%s/.aws-fuzzy/", common.UserHomeDir)
)

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func New(service string) (cachego.Cache, error) {
	if ok, err := exists(CacheDir); !ok {
		err = os.Mkdir(CacheDir, 0700)
		if err != nil {
			return nil, err
		}
	}

	cache := file.New(CacheDir)

	return cache, nil
}
