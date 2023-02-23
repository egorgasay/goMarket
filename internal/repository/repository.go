package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/egorgasay/dockerdb"
	"gomarket/internal/storage"
	"gomarket/internal/storage/postgres"
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

	if cfg.VDB == nil {
		db, err := sql.Open("postgres", cfg.DataSourceCred)
		if err != nil {
			return nil, err
		}

		return postgres.New(db), nil
	}

	cfg.DataSourcePath = "dockerDBs"
	sqlitedb, err := upSqlite(cfg, "DockerDBs-schema.sql")
	if err != nil {
		return nil, err
	}

	stmt, err := sqlitedb.Prepare("SELECT id, connectionString FROM DockerDBs WHERE name = ?")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(cfg.Name)

	err = row.Err()
	if err != nil {
		return nil, err
	}

	err = row.Scan(&cfg.VDB.ID, &cfg.DataSourceCred)
	if err != sql.ErrNoRows && err != nil {
		return nil, err
	}

	ctx := context.TODO()
	err = cfg.VDB.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to run docker storage %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		stmt, err := sqlitedb.Prepare("INSERT INTO DockerDBs VALUES (?, ?, ?)")
		if err != nil {
			return nil, err
		}

		_, err = stmt.Exec(cfg.Name, cfg.VDB.ID, cfg.DataSourceCred)
		if err != nil {
			return nil, err
		}
	}
	sqlitedb.Close()

	return postgres.New(cfg.VDB.DB), nil
}

func upSqlite(cfg *Config, schema string) (*sql.DB, error) {
	exists := storage.IsDBSqliteExist(cfg.DataSourcePath)

	db, err := sql.Open("sqlite3", cfg.DataSourcePath)
	if err != nil {
		return nil, err
	}

	// use migrations instead of this approach
	if !exists {
		err = storage.InitDatabase(db, schema)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
