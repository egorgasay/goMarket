package schema

import (
	"gomarket/internal/loyalty/schema"
	"time"
)

type BalanceMarket struct {
	Current float32 `bson:"balance"`
	Bonuses float32 `bson:"-"`
}

type Customer struct {
	Cookie   string  `bson:"cookie"`
	Login    string  `bson:"login,omitempty" form:"username"`
	Password string  `bson:"password,omitempty" form:"password"`
	Current  float32 `bson:"balance"`
}

type Item struct {
	ID          string  `bson:"_id"`
	Name        string  `bson:"name" form:"name"`
	Price       float32 `bson:"price" form:"price"`
	Description string  `bson:"description,omitempty" form:"desc"`
	Count       int     `bson:"count" form:"qty"`
	ImagePath   string  `bson:"image_path" form:"img"`
}

type Order struct {
	ID     string    `bson:"_id,omitempty"`
	Owner  string    `bson:"owner"`
	Items  []Item    `bson:"items"`
	Date   time.Time `bson:"date"`
	Status string    `bson:"status"`
}

type AuthRequestJSON schema.AuthRequestJSON
type Bonus struct {
	schema.Balance
	Err string `json:"err"`
}

type AccrualRequest struct {
	Order string `json:"order"`
	Goods []Item `json:"goods"`
}
