package storage

import (
	"context"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gomarket/internal/market/schema"
	"log"
)

func (s Storage) CreateUser(login, passwd string) error {
	return nil
}

func (s Storage) CheckPassword(login, passwd string) error {
	return nil
}

func (s Storage) CreateAnonUser(ctx context.Context, user schema.Customer) error {
	c := s.db.Collection("customers")
	_, err := c.InsertOne(ctx, user)
	return err
}

func (s Storage) GetItems(ctx context.Context) ([]schema.Item, error) {
	var items = make([]schema.Item, 0)
	var c = s.db.Collection("items")

	cur, err := c.Find(ctx, bson.D{primitive.E{}})
	if err != nil {
		return nil, err
	}

	for cur.TryNext(ctx) {
		var item schema.Item
		err = cur.Decode(&item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, err
}

func (s Storage) GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error) {
	c := s.db.Collection("customers")

	var filter = bson.D{primitive.E{Key: "cookie", Value: cookie}}
	var balance schema.BalanceMarket
	var err = c.FindOne(ctx, filter).Decode(&balance)

	return balance, err
}

func (s Storage) Buy(ctx context.Context, cookie string, id string) error {
	balance, err := s.GetBalance(ctx, cookie)
	if err != nil {
		return err
	}

	var item schema.Item
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	var filter = bson.D{primitive.E{Key: "_id", Value: ID}}
	c := s.db.Collection("items")

	err = c.FindOne(ctx, filter).Decode(&item)
	if err != nil {
		return err
	}

	if balance.Bonuses+balance.Current < item.Price {
		return ErrNotEnoughMoney
	}

	filter = bson.D{primitive.E{Key: "_id", Value: ID}}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "count", Value: item.Count - 1},
	}}}

	r, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Println(r)

	return nil
}
