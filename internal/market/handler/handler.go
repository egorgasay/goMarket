package handler

import (
	"context"
	"errors"
	echosession "github.com/go-session/echo-session"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/mongo"
	"gomarket/internal/logger"
	"gomarket/internal/market/config"
	"gomarket/internal/market/cookies"
	"gomarket/internal/market/schema"
	"gomarket/internal/market/usecase"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	conf   *config.Config
	logic  usecase.IUseCase // IUseCase for mock tests
	logger logger.ILogger
}

type H map[string]interface{}

const userkey = "user"

func NewHandler(cfg *config.Config, logic usecase.IUseCase, loggerInstance logger.ILogger) *Handler {
	if cfg == nil {
		panic("конфиг равен nil")
	}
	return &Handler{conf: cfg, logic: logic, logger: loggerInstance}
}

func (h Handler) GetMain(c echo.Context) error {
	ctx := context.TODO()
	cookie, _ := c.Cookie("session")
	if cookie == nil {
		var err error
		cookie, err = h.setCookie(ctx, c)
		if err != nil {
			return err
		}
	}

	store := echosession.FromContext(c)
	user, login := store.Get(userkey)
	if !login {
		user = cookie.Value
	}

	username, ok := user.(string)
	if !ok {
		return errors.New("bad username")
	}

	if cart := c.Request().URL.Query().Get("cart"); cart != "" {
		if len(cart) < 2 {
			c.Redirect(http.StatusTemporaryRedirect, "/")
		}

		cart = cart[1:]
		goods := strings.Split(cart, "|")
		err := h.logic.BulkBuy(ctx, cookie.Value, username, h.conf.AccrualSystemAddress, h.conf.LoyaltySystemAddress, goods, login)
		if err != nil {
			h.logger.Warn("Buy:" + err.Error())
			err = c.Render(http.StatusInternalServerError, "check.html", H{})
			if err != nil {
				h.logger.Warn(err.Error())
				return err
			}
			return err
		}

		err = c.Render(http.StatusOK, "check.html", H{"ok": true})
		if err != nil {
			h.logger.Warn(err.Error())
		}
		return err
	}

	items, balance, err := h.getItemsAndBalance(ctx, c, cookie.Value)
	if err != nil {
		err = c.Render(http.StatusInternalServerError, "main_page.html", H{
			"login": login,
			"error": "server error, sorry! we are working on it.",
		})
		if err != nil {
			h.logger.Warn(err.Error())
		}
		return err
	}

	err = c.Render(http.StatusOK, "main_page.html", H{
		"Balance": balance.Current,
		"Bonuses": balance.Bonuses,
		"Items":   reverseSlice(items),
		"login":   login,
	})
	if err != nil {
		h.logger.Warn(err.Error())
	}
	return nil
}

func (h Handler) Login(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		c.Response().WriteHeader(http.StatusOK)
		err := c.Render(http.StatusOK, "auth.html", H{
			"login": true,
		})

		if err != nil {
			h.logger.Warn(err.Error())
		}
		return err
	}

	ctx := context.TODO()
	cookie, _ := c.Cookie("session")
	if cookie == nil {
		var err error
		cookie, err = h.setCookie(ctx, c)
		if err != nil {
			return err
		}
	}

	var user schema.Customer
	err := c.Bind(&user)
	if err != nil {
		h.logger.Warn(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return err
	}

	cookie.Value, err = h.logic.Authentication(user.Login, user.Password)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			h.logger.Warn("user don't exist")
			err = errors.New("user don't exist")
		}
		err = c.Render(http.StatusBadRequest, "auth.html", H{"error": err.Error(), "login": true})
		if err != nil {
			h.logger.Warn(err.Error())
		}
		return err
	}
	c.SetCookie(cookie)

	store := echosession.FromContext(c)
	store.Set(userkey, user.Login)
	err = store.Save()
	if err != nil {
		return err
	}

	c.Redirect(http.StatusTemporaryRedirect, "/")
	return nil
}

func (h Handler) Register(c echo.Context) error {

	if c.Request().Method == http.MethodGet {
		err := c.Render(http.StatusOK, "auth.html", H{})
		if err != nil {
			c.Response().WriteHeader(http.StatusInternalServerError)
			h.logger.Warn(err.Error())
		}
		return err
	}

	ctx := context.TODO()
	cookie, _ := c.Cookie("session")
	if cookie == nil {
		var err error
		_, err = h.setCookie(ctx, c)
		h.logger.Warn(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return err
	}

	var user schema.Customer
	err := c.Bind(&user)
	if err != nil {
		h.logger.Warn(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return err
	}

	newCookie, err := h.logic.CreateUser(user, cookie.Value, h.conf.LoyaltySystemAddress)
	if err != nil {
		h.logger.Warn(err.Error())
		err = c.Render(http.StatusInternalServerError, "auth.html", H{"error": err.Error()})
		if err != nil {
			h.logger.Warn(err.Error())
		}
		return err
	}

	c.SetCookie(cookies.SetCookie(newCookie))

	store := echosession.FromContext(c)
	store.Set(userkey, user.Login)
	err = store.Save()
	if err != nil {
		return err
	}

	c.Redirect(http.StatusTemporaryRedirect, "/")
	return nil
}

func (h Handler) GetOrders(c echo.Context) error {
	ctx := context.TODO()
	cookie, _ := c.Cookie("session")
	if cookie == nil {
		var err error
		cookie, err = h.setCookie(ctx, c)
		if err != nil {
			return err
		}
	}

	store := echosession.FromContext(c)
	user, login := store.Get(userkey)
	if !login {
		user = cookie.Value
	}

	username, ok := user.(string)
	if !ok {
		return errors.New("bad username")
	}

	orders, err := h.logic.GetOrders(ctx, username)
	if err != nil {
		h.logger.Warn(err.Error())
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = c.Render(http.StatusInternalServerError, "orders.html", H{
				"error": "you don't have any orders. go buy smth", "login": login,
			})
			if err != nil {
				h.logger.Warn(err.Error())
			}
			return err
		}

		err = c.Render(http.StatusInternalServerError, "orders.html", H{
			"error": "something went wrong", "login": login,
		})
		if err != nil {
			h.logger.Warn(err.Error())
		}
		return err
	}

	err = c.Render(http.StatusInternalServerError, "orders.html", H{
		"Orders": reverseSlice(orders), "login": login,
	})
	if err != nil {
		h.logger.Warn(err.Error())
	}
	return nil
}

// GetOrderInfo NOW IS UNUSED
func (h Handler) GetOrderInfo(c echo.Context) error {
	ctx := context.TODO()
	cookie, _ := c.Cookie("session")
	if cookie == nil {
		var err error
		cookie, err = h.setCookie(ctx, c)
		if err != nil {
			return err
		}
	}

	store := echosession.FromContext(c)
	user, login := store.Get(userkey)
	if !login {
		user = cookie.Value
	}

	username, ok := user.(string)
	if !ok {
		return errors.New("bad username")
	}

	orderID := c.Request().URL.Query().Get("order")
	if orderID == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/orders")
		return nil
	}

	order, err := h.logic.GetOrder(ctx, username, orderID)
	if err != nil {
		h.logger.Warn(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/orders")
		return nil
	}

	err = c.Render(http.StatusOK, "order_info.html", H{
		"Order": order,
	})

	if err != nil {
		h.logger.Warn(err.Error())
	}

	return nil
}

func (h Handler) GetAdmin(c echo.Context) error {
	ctx := context.TODO()
	store := echosession.FromContext(c)
	user, login := store.Get(userkey)
	if !login {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	username, ok := user.(string)
	if !ok {
		h.logger.Warn("Bad username")
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	isAdmin, err := h.logic.IsAdmin(ctx, username)
	if err != nil || !isAdmin {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return nil
	}

	items, err := h.getItems(ctx, c, "admin.html")
	if err != nil {
		return err
	}

	orders, err := h.getOrders(ctx, c)
	if err != nil {
		return err
	}

	err = c.Render(http.StatusOK, "admin.html", H{
		"Orders": reverseSlice(orders),
		"Items":  reverseSlice(items),
	})

	if err != nil {
		h.logger.Warn(err.Error())
	}

	return nil
}

func (h Handler) PostChangeStatus(c echo.Context) error {
	var ctx = context.TODO()
	store := echosession.FromContext(c)
	user, login := store.Get(userkey)
	if !login {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	username, ok := user.(string)
	if !ok {
		h.logger.Warn("Bad username")
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	isAdmin, err := h.logic.IsAdmin(ctx, username)
	if err != nil || !isAdmin {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return nil
	}

	status := c.FormValue("status")
	statusCode, err := strconv.Atoi(status)
	if err != nil {
		statusCode = 0
	}

	orderID := c.FormValue("orderID")
	err = h.logic.ChangeOrderStatus(ctx, statusCode, orderID)
	if err != nil {
		return h.handleAdminError(ctx, c, err)
	}
	c.Redirect(http.StatusTemporaryRedirect, "/admin")
	return nil
}

func (h Handler) PostAddItem(c echo.Context) error {
	var ctx = context.TODO()
	store := echosession.FromContext(c)
	user, login := store.Get(userkey)
	if !login {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	username, ok := user.(string)
	if !ok {
		h.logger.Warn("Bad username")
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	isAdmin, err := h.logic.IsAdmin(ctx, username)
	if err != nil || !isAdmin {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return nil
	}

	var item schema.Item
	err = c.Bind(&item)
	if err != nil {
		h.logger.Warn(err.Error())
	}

	file, err := c.FormFile("img")
	if err != nil {
		return h.handleAdminError(ctx, c, err)
	}

	item.ImagePath, err = h.saveImage(file)
	if err != nil {
		return h.handleAdminError(ctx, c, err)
	}

	err = h.logic.AddItem(ctx, item)
	if err != nil {
		return h.handleAdminError(ctx, c, err)
	}

	c.Redirect(http.StatusTemporaryRedirect, "/admin")

	return nil
}

func (h Handler) RemoveItem(c echo.Context) error {
	var ctx = context.TODO()
	store := echosession.FromContext(c)
	user, login := store.Get(userkey)
	if !login {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	username, ok := user.(string)
	if !ok {
		h.logger.Warn("Bad username")
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	isAdmin, err := h.logic.IsAdmin(ctx, username)
	if err != nil || !isAdmin {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return nil
	}

	id := c.Request().URL.Query().Get("id")
	if len(id) == 0 {
		return h.handleAdminError(ctx, c, errors.New("empty id"))
	}

	err = h.logic.RemoveItem(ctx, id)
	if err != nil {
		return h.handleAdminError(ctx, c, err)
	}

	c.Redirect(http.StatusTemporaryRedirect, "/admin")

	return nil
}

func (h Handler) ChangeItem(c echo.Context) error {
	var ctx = context.TODO()
	store := echosession.FromContext(c)
	user, login := store.Get(userkey)
	if !login {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	username, ok := user.(string)
	if !ok {
		h.logger.Warn("Bad username")
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	isAdmin, err := h.logic.IsAdmin(ctx, username)
	if err != nil || !isAdmin {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return nil
	}

	ctx = context.Background()
	var item schema.Item
	err = c.Bind(&item)
	if err != nil {
		h.logger.Warn(err.Error())
	}

	file, err := c.FormFile("cimg")
	if err == nil {
		item.ImagePath, err = h.saveImage(file)
		if err != nil {
			return h.handleAdminError(ctx, c, err)
		}
	}

	err = h.logic.ChangeItem(ctx, item)
	if err != nil {
		return h.handleAdminError(ctx, c, err)
	}

	return c.HTML(http.StatusOK, "<p>File uploaded successfully</p><script>window.location.replace(\"/admin\");</script>")
}

func (h Handler) LogoutHandler(c echo.Context) error {
	store := echosession.FromContext(c)
	store.Delete(userkey)
	err := store.Save()
	if err != nil {
		h.logger.Warn("LogoutHandler:" + err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return nil
	}

	c.Redirect(http.StatusTemporaryRedirect, "/login")
	return nil
}
