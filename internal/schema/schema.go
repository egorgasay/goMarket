package schema

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

type ResponseFromTheCalculationSystem struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type WithdrawnRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
