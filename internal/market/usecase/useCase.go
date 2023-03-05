package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ShiraazMoollatjie/goluhn"
	schema2 "gomarket/internal/loyalty/schema"
	"gomarket/internal/market/schema"
	"gomarket/internal/market/storage"
	"io"
	"log"
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

	return uc.storage.CreateUser(user.Login, user.Password, cookie, loyaltyCookie)
}

func (uc UseCase) CheckPassword(login, passwd string) error {
	return uc.storage.CheckPassword(login, passwd)
}

func (uc UseCase) GetBalance(ctx context.Context, cookie string, loyaltyAddress string) (schema.BalanceMarket, error) {
	balance, err := uc.storage.GetBalance(ctx, cookie)
	balance.Bonuses = 0
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+loyaltyAddress+"/api/user/balance", nil)
	if err != nil {
		log.Println(err)
		return balance, nil
	}

	req.Header.Set("Authorization", cookie)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return balance, nil
	}
	defer resp.Body.Close()

	var bonus schema.Bonus
	err = json.NewDecoder(resp.Body).Decode(&bonus)
	if err != nil {
		log.Println(err)
		return balance, nil
	}

	balance.Bonuses = bonus.Current

	return balance, nil
}

func (uc UseCase) CreateAnonUser(ctx context.Context, cookie string) error {
	return uc.storage.CreateAnonUser(ctx, schema.Customer{Cookie: cookie, Current: 10000})
}

func (uc UseCase) GetItems(ctx context.Context) ([]schema.Item, error) {
	return uc.storage.GetItems(ctx)
}

func (uc UseCase) Buy(ctx context.Context, cookie, id, accrualAddress, loyaltyAddress string) error {
	orderID := fmt.Sprint(goluhn.Generate(16)) // TODO: change to 128
	go uc.regNewOrderAccrual(cookie, id, "http://"+accrualAddress+"/api/orders", orderID)
	go uc.regNewOrderLoyalty(cookie, "http://"+loyaltyAddress+"/api/user/orders", orderID)

	balance, err := uc.GetBalance(ctx, cookie, loyaltyAddress)
	if err != nil {
		return err
	}

	item, err := uc.storage.GetItem(ctx, id)
	if err != nil {
		return err
	}

	if balance.Bonuses+balance.Current < item.Price {
		return storage.ErrNotEnoughMoney
	}
	if balance.Bonuses > 0 {
		var amount float32
		if item.Price-balance.Bonuses >= 0 {
			amount = balance.Bonuses
		} else {
			amount = item.Price
		}
		balance.Current = balance.Current + amount
		err = uc.withdrawalBonuses(cookie, orderID, loyaltyAddress, amount)
		if err != nil {
			log.Println("can't write off bonuses:", err)
		}
	}

	return uc.storage.Buy(ctx, cookie, id, balance, item)
}

func (uc UseCase) withdrawalBonuses(cookie, id, loyaltyAddress string, amount float32) error {
	wr := schema2.WithdrawnRequest{
		Order: id,
		Sum:   float64(amount),
	}

	ready, err := json.Marshal(wr)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://"+loyaltyAddress+"/api/user/balance/withdraw", bytes.NewReader(ready))
	if err != nil {
		return err
	}

	return uc.performRequest(req, cookie, http.StatusOK)
}

func (uc UseCase) performRequest(req *http.Request, cookie string, code int) error {
	req.Header.Set("Authorization", cookie)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != code {
		log.Println("Error! Status code:", resp.StatusCode)
		read, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Println("Err text:", string(read))
	}
	return nil
}

func (uc UseCase) regNewOrderAccrual(cookie, id, host, orderID string) {
	item, err := uc.storage.GetItem(context.Background(), id)
	if err != nil {
		log.Println(err)
		return
	}

	item.Description = item.Name

	accrualReq := schema.AccrualRequest{
		Order: orderID,
		Goods: []schema.Item{item},
	}

	ready, err := json.Marshal(accrualReq)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, host, bytes.NewReader(ready))
	if err != nil {
		log.Println(err)
		return
	}

	err = uc.performRequest(req, cookie, http.StatusAccepted)
	if err != nil {
		log.Println(err)
	}
}

func (uc UseCase) regNewOrderLoyalty(cookie, host, orderID string) {
	req, err := http.NewRequest(http.MethodPost, host, strings.NewReader(orderID))
	if err != nil {
		return
	}

	err = uc.performRequest(req, cookie, http.StatusAccepted)
	if err != nil {
		log.Println(err)
	}
}
