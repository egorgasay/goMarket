package schema

import "time"

type AuthRequestJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserOrder struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

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

type ResponseFromTheCalculationSystem struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type WithdrawnRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdrawn struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
