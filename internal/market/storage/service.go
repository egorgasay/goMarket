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
	CreateUser(login, passwd, cookie string, newCookie string) (string, error)
	Authentication(login, passwd string) (string, error)
	GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error)
	GetItems(ctx context.Context) ([]schema.Item, error)
	GetItem(ctx context.Context, id string) (schema.Item, error)
	Buy(ctx context.Context, cookie, id string, balance schema.BalanceMarket, item schema.Item) error
	GetOrders(ctx context.Context, cookie string) ([]schema.Order, error)
	GetAllOrders(ctx context.Context) ([]schema.Order, error)
	AddOrder(ctx context.Context, order schema.Order) error
	AddItem(ctx context.Context, item schema.Item) error
	RemoveItem(ctx context.Context, id string) error
	ChangeItem(ctx context.Context, item schema.Item) error
	IsAdmin(ctx context.Context, username string) (bool, error)
	ChangeOrderStatus(ctx context.Context, status schema.Status, orderID string) error
	GetOrder(ctx context.Context, username, orderID string) (order schema.Order, err error)
}

type Storage struct {
	db *mongo.Database
}

type Type string

var ErrUsernameConflict = errors.New("username already exists")
var ErrWrongPassword = errors.New("wrong password")
var ErrBadCookie = errors.New("bad cookie")
var ErrNotEnoughMoney = errors.New("insufficient funds for payment")

type Config struct {
	DriverName     Type
	DataSourceCred string
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
		db: client.Database("test"),
	}

	return dao, nil
}
