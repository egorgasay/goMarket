package repository

import (
	"database/sql"
	"github.com/egorgasay/dockerdb"
	_ "github.com/mattn/go-sqlite3"
	"gomarket/internal/loyalty/storage"
)

type Config struct {
	DriverName     storage.Type
	DataSourceCred string
	DataSourcePath string
	VDB            *dockerdb.VDB
	Name           string
}

func New(cfg *Config) (storage.IStorage, error) {
	if cfg == nil {
		panic("конфигурация задана некорректно")
	}

	db, err := sql.Open("postgres", cfg.DataSourceCred)
	if err != nil {
		return nil, err
	}
	return storage.New(db, "file://internal/loyalty/storage/migrations"), nil
}
