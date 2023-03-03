package storage

import (
	"context"
	"errors"
	"gomarket/internal/market/schema"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type IStorage interface {
	CreateAnonUser(ctx context.Context, user schema.Customer) error
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error)
	GetItems(ctx context.Context) ([]schema.Item, error)
	Buy(ctx context.Context, cookie string, id string) error
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
		db: client.Database("test"),
	}
  
  //var item = schema.Item{Name: "MYSTERY BOX", Price: 500, ImagePath: "logo.png", Count: 100}
 //_, err = dao.db.Collection("items").InsertOne(ctx, item)
  //log.Println(err)
  
	return dao, nil
}
