package cache

import (
	"database/sql"
	"fmt"
	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/faabiosr/cachego"
	"github.com/faabiosr/cachego/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var (
	cacheDir = fmt.Sprintf("%s/.aws-fuzzy/", common.UserHomeDir)
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
	if ok, err := exists(cacheDir); !ok {
		err = os.Mkdir(cacheDir, 0700)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/cache.sqlite", cacheDir))
	if err != nil {
		fmt.Printf("failed to open cache, %s\n", err)
		return nil, err
	}

	cache, err := sqlite3.New(db, service)
	if err != nil {
		fmt.Printf("failed to create sqlite db, %s\n", err)
		return cache, err
	}

	return cache, nil
}
