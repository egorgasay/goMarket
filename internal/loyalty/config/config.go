package config

import (
	"flag"
	"gomarket/internal/loyalty/cookies"
	"gomarket/internal/loyalty/storage"
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
	Host                 string
	Key                  []byte
	DBConfig             *storage.Config
	AccrualSystemAddress string
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

	if *f.dsn == "" {
		log.Println("Here!!")
	}

	return &Config{
		Host: *f.host,
		Key:  []byte("CHANGE ME"),
		DBConfig: &storage.Config{
			DriverName:     "postgres",
			DataSourceCred: *f.dsn,
			Name:           "vdb",
		},
		AccrualSystemAddress: *f.asa,
	}
}
