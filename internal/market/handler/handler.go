package handler

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"gomarket/internal/logger"
	"gomarket/internal/market/config"
	"gomarket/internal/market/cookies"
	"gomarket/internal/market/schema"
	"gomarket/internal/market/usecase"
	"net/http"

	"github.com/labstack/echo"
)

type Handler struct {
	conf   *config.Config
	logic  usecase.IUseCase // IUseCase for mock tests
	logger logger.ILogger
}

type H map[string]interface{}

func NewHandler(cfg *config.Config, logic usecase.IUseCase, loggerInstance zerolog.Logger) *Handler {
	if cfg == nil {
		panic("конфиг равен nil")
	}
	return &Handler{conf: cfg, logic: logic, logger: logger.New(loggerInstance)}
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

	balance, err := h.logic.GetBalance(ctx, cookie)
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
		err = c.Render(http.StatusOK, "main_page.html", H{"error": err})
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

	if id := c.Request().URL.Query().Get("id"); id != "" {
		err := h.logic.Buy(ctx, cookie.Value, id)
		if err != nil {
			h.logger.Warn("Buy:" + err.Error())
			err = c.Render(http.StatusOK, "main_page.html", H{"error": err})
			if err != nil {
				h.logger.Warn(err.Error())
				return err
			}
			return err
		}

		items, balance, err := h.getItemsAndBalance(ctx, c, cookie.Value)
		if err != nil {
			return err
		}

		err = c.Render(http.StatusOK, "main_page.html", H{
			"Balance": balance.Current,
			"Bonuses": balance.Bonuses,
			"Items":   items,
			"Bought":  true,
		})
		if err != nil {
			h.logger.Warn(err.Error())
		}
		return err
	}

	items, balance, err := h.getItemsAndBalance(ctx, c, cookie.Value)
	if err != nil {
		return err
	}

	err = c.Render(http.StatusOK, "main_page.html", H{
		"Balance": balance.Current,
		"Bonuses": balance.Bonuses,
		"Items":   items,
	})
	if err != nil {
		h.logger.Warn(err.Error())
	}
	return err
}
