package storage

import (
	"context"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gomarket/internal/market/schema"
)

func (s Storage) CreateUser(login, passwd string) error {
	return nil
}

func (s Storage) CheckPassword(login, passwd string) error {
	return nil
}

func (s Storage) CreateAnonUser(ctx context.Context, user schema.Customer) error {
	_, err := s.c.InsertOne(ctx, user)
	return err
}

func (s Storage) GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error) {
	var filter = bson.D{primitive.E{Key: "cookie", Value: cookie}}
	var balance schema.BalanceMarket
	var err = s.c.FindOne(ctx, filter).Decode(&balance)
	return balance, err
}
