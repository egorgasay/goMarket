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

func (h Handler) getItemsAndBalance(ctx context.Context, c echo.Context, cookie string) ([]schema.Item, schema.BalanceMarket, error) {
	items, err := h.logic.GetItems(ctx)
	if err != nil {
		err := c.Render(http.StatusOK, "main_page.html", H{"error": err})
		if err != nil {
			h.logger.Warn("GetItems" + err.Error())
		}
		return nil, schema.BalanceMarket{}, err
	}

	balance, err := h.logic.GetBalance(ctx, cookie, h.conf.LoyaltySystemAddress)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, schema.BalanceMarket{}, err
		}
		_, err := h.setCookie(ctx, c)
		if err != nil {
			return nil, schema.BalanceMarket{}, err
		}

		c.Redirect(http.StatusTemporaryRedirect, "/")
		return nil, schema.BalanceMarket{}, err
	}

	return items, balance, nil
}

func (h Handler) setCookie(ctx context.Context, c echo.Context) (*http.Cookie, error) {
	cookie := cookies.SetCookie()
	c.SetCookie(cookie)

	err := h.logic.CreateAnonUser(ctx, cookie.Value)
	if err != nil {
		h.logger.Warn("ErrCreateAnonUser:" + err.Error())
		err = c.Render(http.StatusInternalServerError, "main_page.html", H{"error": err})
		if err != nil {
			h.logger.Warn(err.Error())
			return nil, err
		}
	}

	return cookie, nil
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
	_, login := store.Get(userkey)

	if id := c.Request().URL.Query().Get("id"); id != "" {
		err := h.logic.Buy(ctx, cookie.Value, id, h.conf.AccrualSystemAddress, h.conf.LoyaltySystemAddress)
		if err != nil {
			h.logger.Warn("Buy:" + err.Error())
			err = c.Render(http.StatusInternalServerError, "main_page.html", H{"error": err})
			if err != nil {
				h.logger.Warn(err.Error())
				return err
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
			"Items":   items,
			"Bought":  true,
			"login":   login,
		})
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
		"Items":   items,
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
		_, err = h.setCookie(ctx, c)
		if err != nil {
			return err
		}
	}

	username := c.Request().PostForm.Get("username")
	password := c.Request().PostForm.Get("password")

	err := h.logic.CheckPassword(username, password)
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
