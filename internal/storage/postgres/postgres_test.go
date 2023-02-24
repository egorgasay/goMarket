package postgres

import (
	"context"
	"github.com/egorgasay/dockerdb"
	"log"
	"os"
	"testing"
)

var TestDB Postgres

func TestMain(m *testing.M) {
	// Write code here to run before tests
	ctx := context.TODO()
	cfg := dockerdb.CustomDB{
		DB: dockerdb.DB{
			Name:     "vdb_test",
			User:     "admin",
			Password: "admin",
		},
		Port: "1254",
		Vendor: dockerdb.Vendor{
			Name:  dockerdb.Postgres,
			Image: "postgres:15", // TODO: add dockerdb.Postgres15 as image into dockerdb package
		},
	}
	vdb, err := dockerdb.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
		return
	}

	TestDB = Postgres{vdb.DB}
	// Run tests
	exitVal := m.Run()

	// Write code here to run after tests
	queries := []string{
		"DROP SCHEMA public CASCADE;",
		"CREATE SCHEMA public;",
		"GRANT ALL ON SCHEMA public TO postgres;",
		"GRANT ALL ON SCHEMA public TO public;",
		"COMMENT ON SCHEMA public IS 'standard public schema';",
	}

	tx, err := TestDB.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	for _, query := range queries {
		_, err := tx.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	// Exit with exit value from tests
	os.Exit(exitVal)
}

func TestPostgres_CreateUser(t *testing.T) {
	login := "admin"
	password := "admin"
	err := TestDB.CreateUser(login, password)
	if err != nil {
		t.Fatal(err)
	}

	var username string
	err = TestDB.DB.QueryRow(
		`SELECT "Name" FROM "Users" WHERE "Password" = $1`, password).
		Scan(&username)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostgres_CheckPassword(t *testing.T) {
	login := "admin"
	password := "admin"
	err := TestDB.CheckPassword(login, password)
	if err != nil {
		t.Fatal(err)
	}

	var username string
	err = TestDB.DB.QueryRow(
		`SELECT "Name" FROM "Users" WHERE "Password" = $1`, password).
		Scan(&username)
	if err != nil {
		t.Fatal(err)
	}
}
