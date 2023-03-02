package storage

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"gomarket/internal/loyalty/schema"
	"log"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type IStorage interface {
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	CheckID(username, id string) error
	GetOrders(username string) (Orders, error)
	GetBalance(username string) (schema.Balance, error)
	UpdateOrder(username, id, status string, accrual float64) error
	Withdraw(username string, amount float64, orderID string) error
	GetWithdrawals(username string) ([]schema.Withdrawn, error)
}

type Storage struct {
	DB *sql.DB
}

type Type string

type Orders []schema.UserOrder

var ErrUsernameConflict = errors.New("username already exists")
var ErrWrongPassword = errors.New("wrong password")
var ErrBadCookie = errors.New("bad cookie")
var ErrCreatedByAnotherUser = errors.New("uid already exists and created by another user")
var ErrCreatedByThisUser = errors.New("uid already exists and created by this user")
var ErrBadID = errors.New("wrong id format")
var ErrNoResult = errors.New("the user has no orders")
var ErrNotEnoughMoney = errors.New("insufficient funds for payment")
var ErrNoWithdrawals = errors.New("user don't have withdrawals operations")

//var ErrWrongOrderID = errors.New("wrong order id")

type Config struct {
	DriverName     Type
	DataSourceCred string
	DataSourcePath string
	Name           string
}

func Init(cfg *Config) (IStorage, error) {
	if cfg == nil {
		panic("конфигурация задана некорректно")
	}

	db, err := sql.Open("postgres", cfg.DataSourceCred)
	if err != nil {
		return nil, err
	}

	return New(db, "file://internal/loyalty/storage/migrations"), nil
}

func New(db *sql.DB, pathToMigrations string) IStorage {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	m, err := migrate.NewWithDatabaseInstance(
		pathToMigrations,
		"gomarket", driver)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = m.Up()
	if err != nil {
		if err.Error() != "no change" {
			log.Fatal(err)
		}
	}

	return Storage{DB: db}
}
