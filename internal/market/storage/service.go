package storage

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gomarket/internal/market/schema"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type IStorage interface {
	CreateAnonUser(ctx context.Context, user schema.Customer) error
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error)
}

type Storage struct {
	c *mongo.Collection
}

type Type string

var ErrUsernameConflict = errors.New("username already exists")
var ErrWrongPassword = errors.New("wrong password")
var ErrBadCookie = errors.New("bad cookie")
var ErrNotEnoughMoney = errors.New("insufficient funds for payment")

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
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DataSourceCred))
	if err != nil {
		return nil, err
	}

	dao := &Storage{
		c: client.Database("core").Collection("customers"),
	}

	return dao, nil
}
