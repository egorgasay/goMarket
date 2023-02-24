package config

import (
	"context"
	"flag"
	"github.com/egorgasay/dockerdb"
	"gomarket/internal/cookies"
	"gomarket/internal/repository"
	"log"
	"os"
)

const (
	defaultHost = ":8080"
)

type Flag struct {
	host *string
	dsn  *string
	asa  *string
	key  string
}

var f Flag

func init() {
	f.host = flag.String("a", defaultHost, "-a=host")
	f.dsn = flag.String("d", "", "-d=connection_string")
	f.asa = flag.String("r", "", "-r=host")
}

type Config struct {
	Host     string
	BaseURL  string
	Key      []byte
	DBConfig *repository.Config
}

func New() *Config {
	flag.Parse()

	if addr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		f.host = &addr
	}

	if asa, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		f.asa = &asa
	}

	if dsn, ok := os.LookupEnv("DATABASE_URI"); ok {
		f.dsn = &dsn
	}

	if key, ok := os.LookupEnv("KEY"); ok {
		f.key = key
		cookies.SetSecret([]byte(key))
	}

	var ddb *dockerdb.VDB

	if *f.dsn == "" {
		ctx := context.TODO()

		cfg := dockerdb.CustomDB{
			DB: dockerdb.DB{
				Name:     "vdb2",
				User:     "admin",
				Password: "admin",
			},
			Port: "12589",
			Vendor: dockerdb.Vendor{
				Name:  dockerdb.Postgres,
				Image: "postgres", // TODO: add dockerdb.Postgres15 as image into dockerdb package
			},
		}

		var err error
		ddb, err = dockerdb.New(ctx, cfg)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &Config{
		Host: *f.host,
		Key:  []byte("CHANGE ME"),
		DBConfig: &repository.Config{
			DriverName:     "postgres",
			DataSourceCred: *f.dsn,
			VDB:            ddb,
			Name:           "vdb",
		},
	}
}
