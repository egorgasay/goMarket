package config

import (
	"flag"
	"gomarket/internal/loyalty/cookies"
	"gomarket/internal/market/storage"
	"os"
)

const (
	defaultHost = ":8080"
)

type Flag struct {
	host    *string
	dsn     *string
	asa     *string
	loyalty *string
	key     string
}

var f Flag

func init() {
	f.host = flag.String("a", defaultHost, "-a=host")
	f.dsn = flag.String("d", "", "-d=connection_string")
	f.asa = flag.String("r", "http://127.0.0.1:8070", "-r=host")
	f.loyalty = flag.String("l", "http://127.0.0.1:8000", "-l=host")
}

type Config struct {
	Host                 string
	Key                  []byte
	DBConfig             *storage.Config
	AccrualSystemAddress string
	LoyaltySystemAddress string
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

	if loyalty, ok := os.LookupEnv("LOYALTY"); ok {
		f.loyalty = &loyalty
	}

	if key, ok := os.LookupEnv("KEY"); ok {
		f.key = key
		cookies.SetSecret([]byte(key))
	}

	return &Config{
		Host: *f.host,
		Key:  []byte("CHANGE ME"),
		DBConfig: &storage.Config{
			DriverName:     "mongo",
			DataSourceCred: *f.dsn,
		},
		AccrualSystemAddress: *f.asa,
		LoyaltySystemAddress: *f.loyalty,
	}
}
