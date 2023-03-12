package handler

import (
	"context"
	"errors"
	"github.com/docker/distribution/uuid"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/mongo"
	"gomarket/internal/market/cookies"
	"gomarket/internal/market/schema"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func (h Handler) getItemsAndBalance(ctx context.Context, c echo.Context, cookie string) ([]schema.Item, schema.BalanceMarket, error) {
	items, err := h.getItems(ctx, c, "main_page.html")
	if err != nil {
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

func (h Handler) getItems(ctx context.Context, c echo.Context, template string) ([]schema.Item, error) {
	items, err := h.logic.GetItems(ctx)
	if err != nil {
		h.logger.Warn(err.Error())
		err = c.Render(http.StatusOK, template, H{
			"error":  err,
			"Orders": []schema.Order{},
			"Admin":  template == "admin.html"})
		if err != nil {
			h.logger.Warn("GetItems" + err.Error())
		}
		return items, err
	}
	return items, nil
}

func (h Handler) getOrders(ctx context.Context, c echo.Context) ([]schema.Order, error) {
	orders, err := h.logic.GetAllOrders(ctx)
	if err != nil {
		err = c.Render(http.StatusInternalServerError, "admin.html", H{
			"error": err, "Orders": orders,
		})
		return orders, err
	}
	return orders, nil
}

func (h Handler) handleAdminError(ctx context.Context, c echo.Context, err error) error {
	h.logger.Warn(err.Error())

	items, err := h.getItems(ctx, c, "admin.html")
	if err != nil {
		return err
	}

	orders, err := h.getOrders(ctx, c)
	if err != nil {
		return err
	}

	err = c.Render(http.StatusOK, "admin.html", H{
		"error":  err,
		"Orders": orders,
		"Items":  items,
	})
	if err != nil {
		h.logger.Warn("GetItems" + err.Error())
	}
	return err
}

func (h Handler) saveImage(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		h.logger.Warn(err.Error())
		return "", err
	}
	defer src.Close()

	split := strings.Split(file.Filename, ".")
	if len(split) > 1 {
		file.Filename = uuid.Generate().String() + "." + split[len(split)-1]
	} else {
		file.Filename = uuid.Generate().String()
	}

	fileName := "static/img/" + file.Filename
	// Destination
	dst, err := os.Create(fileName)
	if err != nil {
		h.logger.Warn(err.Error())
		return "", err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return fileName, nil
}

type Slices interface {
	schema.Order | schema.Item
}

func reverseSlice[S Slices](sl []S) []S {
	for i, j := 0, len(sl)-1; i < j; i, j = i+1, j-1 {
		sl[i], sl[j] = sl[j], sl[i]
	}

	return sl
}
