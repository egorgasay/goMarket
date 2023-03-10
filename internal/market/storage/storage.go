package storage

import (
	"context"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gomarket/internal/market/schema"
)

func (s Storage) CreateUser(login, passwd, cookie string, newCookie string) (string, error) {
	c := s.db.Collection("customers")
	filter := bson.M{"cookie": cookie}

	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "login", Value: login},
		primitive.E{Key: "password", Value: passwd},
		primitive.E{Key: "cookie", Value: newCookie},
	}}}
	option := options.FindOneAndUpdate()
	ctx := context.TODO()
	c.FindOneAndUpdate(ctx, filter, update, option)

	c = s.db.Collection("orders")
	filter = bson.M{"owner": cookie}
	update = bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "owner", Value: login}}}}
	_, err := c.UpdateMany(ctx, filter, update)
	if err != nil {
		return "", err
	}

	return newCookie, nil
}

func (s Storage) Authentication(login, passwd string) (string, error) {
	c := s.db.Collection("customers")
	var filter = bson.D{primitive.E{Key: "login", Value: login}}

	ctx := context.TODO()
	var user schema.Customer
	err := c.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return "", err
	}

	if user.Password != passwd {
		return "", ErrWrongPassword
	}

	return user.Cookie, nil
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

func (s Storage) GetItem(ctx context.Context, id string) (schema.Item, error) {
	c := s.db.Collection("items")
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return schema.Item{}, err
	}

	var filter = bson.D{primitive.E{Key: "_id", Value: ID}}
	var item schema.Item
	err = c.FindOne(ctx, filter).Decode(&item)

	return item, err
}

func (s Storage) GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error) {
	c := s.db.Collection("customers")

	var filter = bson.D{primitive.E{Key: "cookie", Value: cookie}}
	var balance schema.BalanceMarket
	var err = c.FindOne(ctx, filter).Decode(&balance)

	return balance, err
}

func (s Storage) Buy(ctx context.Context, cookie, id string, balance schema.BalanceMarket, item schema.Item) error {
	c := s.db.Collection("items")
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: ID}}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "count", Value: item.Count},
	}}}

	_, err = c.UpdateOne(ctx, filter, update)
	if err != nil {
		// withdrawal rollback (implement deposit handler)
		return err
	}

	c = s.db.Collection("customers")

	filter = bson.D{primitive.E{Key: "cookie", Value: cookie}}
	update = bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "balance", Value: balance.Current - item.Price},
	}}}

	_, err = c.UpdateOne(ctx, filter, update)
	return err
}

func (s Storage) GetOrders(ctx context.Context, cookie string) ([]schema.Order, error) {
	c := s.db.Collection("orders")
	filter := bson.D{primitive.E{Key: "owner", Value: cookie}}
	cur, err := c.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	orders := make([]schema.Order, 0)

	for cur.TryNext(ctx) {
		var order schema.Order
		err = cur.Decode(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return orders, nil
}

func (s Storage) GetAllOrders(ctx context.Context) ([]schema.Order, error) {
	c := s.db.Collection("orders")

	orders := make([]schema.Order, 0)
	filter := bson.D{primitive.E{}}
	cur, err := c.Find(ctx, filter)
	if err != nil {
		return orders, err
	}

	for cur.TryNext(ctx) {
		var order schema.Order
		err = cur.Decode(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return orders, mongo.ErrNoDocuments
	}

	return orders, nil
}

func (s Storage) AddOrder(ctx context.Context, order schema.Order) error {
	c := s.db.Collection("orders")
	_, err := c.InsertOne(ctx, order)
	return err
}

func (s Storage) AddItem(ctx context.Context, item schema.Item) error {
	c := s.db.Collection("items")
	_, err := c.InsertOne(ctx, item)
	return err
}

func (s Storage) RemoveItem(ctx context.Context, id string) error {
	c := s.db.Collection("items")
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.D{primitive.E{Key: "_id", Value: ID}}
	_, err = c.DeleteOne(ctx, filter)
	return err
}

func (s Storage) ChangeItem(ctx context.Context, item schema.Item) error {
	c := s.db.Collection("items")
	ID, err := primitive.ObjectIDFromHex(item.ID)
	if err != nil {
		return err
	}

	item.ID = ""
	filter := bson.D{primitive.E{Key: "_id", Value: ID}}
	_, err = c.ReplaceOne(ctx, filter, item)
	return err
}
