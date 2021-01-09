package cache

import (
	"database/sql"
	"fmt"
	"github.com/faabiosr/cachego"
	"github.com/faabiosr/cachego/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func New(service string) (cachego.Cache, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/.aws-fuzzy/cache.sqlite", os.Getenv("HOME")))
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
