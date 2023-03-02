package schema

import "time"

type BalanceMarket struct {
	Current float32 `bson:"balance"`
	Bonuses float32 `bson:"-"`
}

type Customer struct {
	Cookie   string  `bson:"cookie"`
	Login    string  `bson:"login,omitempty"`
	Password string  `bson:"password,omitempty"`
	Current  float32 `bson:"balance"`
}

type Item struct {
	ID          string  `bson:"_id"`
	Name        string  `bson:"name"`
	Price       float32 `bson:"price"`
	Description string  `bson:"description,omitempty"`
	Count       int     `bson:"count"`
	ImagePath   string  `bson:"image_path"`
}

type Order struct {
	ID    string    `bson:"ID"`
	Items []Item    `bson:"items"`
	Date  time.Time `bson:"date"`
}
