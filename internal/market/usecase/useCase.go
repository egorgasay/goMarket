package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gomarket/internal/loyalty/storage"
	"gomarket/internal/market/schema"
	"net/http"
	"strings"
)

var ErrReservedUsername = errors.New("username is reserved")
var ErrServer = errors.New("server error, sorry! we're already working on it")
var ErrBadCookie = errors.New("bad cookie")

func (uc UseCase) CreateUser(user schema.Customer, cookie string, loyaltyAddress string) (string, error) {
	jsonMSG, err := json.Marshal(&user)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(jsonMSG)
	resp, err := http.Post("http://"+loyaltyAddress+"/api/user/register", "application/json", reader)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if http.StatusConflict == resp.StatusCode {
		return "", ErrReservedUsername
	} else if resp.StatusCode != http.StatusOK {
		return "", ErrServer
	}

	loyaltyCookie := resp.Header.Get("Authorization")
	if loyaltyCookie == "" {
		return "", ErrBadCookie
	}

	split := strings.Split(loyaltyCookie, "-")
	if len(split) != 2 {
		return "", storage.ErrBadCookie
	}

	return uc.storage.CreateUser(user.Login, user.Password, cookie, split[1])
}

func (uc UseCase) CheckPassword(login, passwd string) error {
	return uc.storage.CheckPassword(login, passwd)
}

func (uc UseCase) GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error) {
	balance, err := uc.storage.GetBalance(ctx, cookie)
	balance.Bonuses = 0 // TODO: Connect with loyalty service
	return balance, err
}

func (uc UseCase) CreateAnonUser(ctx context.Context, cookie string) error {
	return uc.storage.CreateAnonUser(ctx, schema.Customer{Cookie: cookie, Current: 10000})
}

func (uc UseCase) GetItems(ctx context.Context) ([]schema.Item, error) {
	return uc.storage.GetItems(ctx)
}

func (uc UseCase) Buy(ctx context.Context, cookie string, id string) error {
	// go RegNewAccural
	// go RegNewLoyalty

	return uc.storage.Buy(ctx, cookie, id)
}

func regNewOrder(ctx context.Context, cookie string, id string) error {
	return nil
}
