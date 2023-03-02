package repository

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"gomarket/internal/storage"
)

type Config struct {
	DriverName     storage.Type
	DataSourceCred string
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

	return storage.New(db, "file://internal/storage/migrations"), nil
}
