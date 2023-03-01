package handler

import (
	"context"
	"github.com/labstack/echo"
	"gomarket/internal/cookies"
	"gomarket/internal/market/config"
	"gomarket/internal/market/usecase"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Handler struct {
	conf  *config.Config
	logic usecase.IUseCase // IUseCase for mock tests
}

type H map[string]interface{}

var lastUID int64 = 0
var idMu = &sync.Mutex{}

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
		log.Println("Setting new cookie...")
		idMu.Lock()
		lastUID++
		var uid = lastUID
		idMu.Unlock()

		cookie = cookies.NewCookie(strconv.FormatInt(uid, 10))
		c.SetCookie(cookie)

		err := h.logic.CreateAnonUser(ctx, cookie.Value)
		log.Println(err)
		err = c.Render(http.StatusOK, "main_page.html", H{"error": err})
		if err != nil {
			log.Println(err)
			return err
		}
	}

	balance, err := h.logic.GetBalance(ctx, cookie.Value)
	if err != nil {
		err := c.Render(http.StatusOK, "main_page.html", H{"error": err})
		if err != nil {
			log.Println(err)
		}
		return err
	}

	err = c.Render(http.StatusOK, "main_page.html", H{"Balance": balance.Current,
		"Bonuses": balance.Bonuses})
	if err != nil {
		log.Println(err)
	}
	return err
}
