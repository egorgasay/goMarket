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
	Current   string `json:"current"`
	Withdrawn string `json:"withdrawn"`
}

type ResponseFromTheCalculationSystem struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
