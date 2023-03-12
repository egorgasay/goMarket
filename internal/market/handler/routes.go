package handler

import (
	"github.com/labstack/echo"
)

func (h Handler) PublicRoutes(e *echo.Echo) {
	e.Any("/", h.GetMain)
	e.Any("/login", h.Login)
	e.Any("/reg", h.Register)
	e.GET("/orders", h.GetOrders)
	e.GET("/order", h.GetOrderInfo)
}

func (h Handler) PrivateRoutes(e *echo.Echo) {
	g := e.Group("/admin")
	g.Any("/", h.GetAdmin)
	g.POST("/add-item", h.PostAddItem)
	g.GET("/remove", h.RemoveItem)
	g.POST("/change", h.ChangeItem)
	g.POST("/change-status", h.PostChangeStatus)
}
