package schema

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
