package handler

import (
	"context"
	"github.com/labstack/echo"
	"gomarket/internal/market/config"
	"gomarket/internal/market/cookies"
	"gomarket/internal/market/usecase"
	"log"
	"net/http"
)

type Handler struct {
	conf  *config.Config
	logic usecase.IUseCase // IUseCase for mock tests
}

type H map[string]interface{}

func NewHandler(cfg *config.Config, logic usecase.IUseCase) *Handler {
	if cfg == nil {
		panic("конфиг равен nil")
	}

	return &Handler{conf: cfg, logic: logic}
}

func (h Handler) GetMain(c echo.Context) error {
	ctx := context.TODO()
	cookie, _ := c.Cookie("session")
	if cookie == nil {
		cookie = cookies.SetCookie()
		c.SetCookie(cookie)
		log.Println(cookie.Value)

		err := h.logic.CreateAnonUser(ctx, cookie.Value)
		if err != nil {
			log.Println("ErrCreateAnonUser:", err)
			err = c.Render(http.StatusOK, "main_page.html", H{"error": err})
			if err != nil {
				log.Println(err)
				return err
			}
			return nil
		}
	}

	items, err := h.logic.GetItems(ctx)
	if err != nil {
		err := c.Render(http.StatusOK, "main_page.html", H{"error": err})
		if err != nil {
			log.Println("GetItems:", err)
		}
		return err
	}

	if id := c.Request().URL.Query().Get("id"); id != "" {
		err = h.logic.Buy(ctx, cookie.Value, id)
		if err != nil {
			log.Println("Buy", err)
			err = c.Render(http.StatusOK, "main_page.html", H{"error": err})
			if err != nil {
				log.Println(err)
				return err
			}
			return err
		}

		items, err = h.logic.GetItems(ctx)
		if err != nil {
			err := c.Render(http.StatusOK, "main_page.html", H{"error": err})
			if err != nil {
				log.Println("GetItems:", err)
			}
			return err
		}

		balance, err := h.logic.GetBalance(ctx, cookie.Value)
		if err != nil {
			err := c.Render(http.StatusOK, "main_page.html", H{"error": err})
			if err != nil {
				log.Println("GetBalance:", err)
			}
			return err
		}

		err = c.Render(http.StatusOK, "main_page.html", H{
			"Balance": balance.Current,
			"Bonuses": balance.Bonuses,
			"Items":   items,
		})
		if err != nil {
			log.Println(err)
		}
		return err
	}

	balance, err := h.logic.GetBalance(ctx, cookie.Value)
	if err != nil {
		err := c.Render(http.StatusOK, "main_page.html", H{"error": err})
		if err != nil {
			log.Println("GetBalance:", err)
		}
		return err
	}

	items, err = h.logic.GetItems(ctx)
	if err != nil {
		err := c.Render(http.StatusOK, "main_page.html", H{"error": err})
		if err != nil {
			log.Println("GetItems:", err)
		}
		return err
	}

	err = c.Render(http.StatusOK, "main_page.html", H{
		"Balance": balance.Current,
		"Bonuses": balance.Bonuses,
		"Items":   items,
	})
	if err != nil {
		log.Println(err)
	}
	return err
}

//func (h Handler) IDInterceptor(c echo.Context) error {
//	c.Request().URL.
//}
