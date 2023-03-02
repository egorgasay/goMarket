package handler

import (
	"context"
	"github.com/labstack/echo"
	"gomarket/internal/market/config"
	"gomarket/internal/market/usecase"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
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

		fid := strconv.FormatInt(uid, 10)
		cookie = new(http.Cookie)
		cookie.Name = "session"
		cookie.Value = fid
		cookie.Expires = time.Now().Add(24 * time.Hour * 365)
		c.SetCookie(cookie)
		log.Println("cookie:", cookie)

		err := h.logic.CreateAnonUser(ctx, cookie.Value)
		if err != nil {
			log.Println("mongo:", err)
			err = c.Render(http.StatusOK, "main_page.html", H{"error": err})
			if err != nil {
				log.Println(err)
				return err
			}
			return nil
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
