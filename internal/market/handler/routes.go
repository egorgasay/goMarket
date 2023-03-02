package handler

import (
	"github.com/go-chi/chi"
	"gomarket/internal/middleware"
)

//func (h Handler) PublicRoutes(r chi.Router) {
//	r.Get("/", h.GetMain())
//}

func (h Handler) PrivateRoutes(r chi.Router) {
	r.Use(middleware.AuthRequired)
	//r.Post("/api/user/orders", h.PostOrders())
	//r.Get("/api/user/orders", h.GetUserOrders())
	//
	//r.Get("/api/user/balance", h.GetBalance())
	//
	//r.Post("/api/user/balance/withdraw", h.PostWithdraw())
	//r.Get("/api/user/withdrawals", h.GetWithdrawals())
}
