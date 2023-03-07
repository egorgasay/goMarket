package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ShiraazMoollatjie/goluhn"
	schema2 "gomarket/internal/loyalty/schema"
	"gomarket/internal/market/schema"
	"gomarket/internal/market/storage"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func (uc UseCase) CreateUser(user schema.Customer, cookie string, loyaltyAddress string) (string, error) {
	jsonMSG, err := json.Marshal(&user)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(jsonMSG)
	resp, err := http.Post("http://"+loyaltyAddress+"/api/user/register", "application/json", reader)
	if err != nil {
		log.Println(err)
		return "", ErrDeadLoyalty
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

func (uc UseCase) Authentication(login, passwd string) (string, error) {
	return uc.storage.Authentication(login, passwd)
}

func (uc UseCase) GetBalance(ctx context.Context, cookie string, loyaltyAddress string) (schema.BalanceMarket, error) {
	balance, err := uc.storage.GetBalance(ctx, cookie)
	if err != nil {
		log.Println(err)
		return balance, nil
	}

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

func (uc UseCase) Buy(ctx context.Context, cookie, id, accrualAddress, loyaltyAddress string, count int, login bool) error {
	orderID := fmt.Sprint(goluhn.Generate(16)) // TODO: change to 128

	balance, err := uc.GetBalance(ctx, cookie, loyaltyAddress)
	if err != nil {
		return err
	}

	item, err := uc.storage.GetItem(ctx, id)
	if err != nil {
		return err
	}

	item.Price = float32(count) * item.Price

	if balance.Bonuses+balance.Current < item.Price {
		return storage.ErrNotEnoughMoney
	}

	if login {
		go uc.regNewOrderAccrual(cookie, id, "http://"+accrualAddress+"/api/orders", orderID, count)
		go uc.regNewOrderLoyalty(cookie, "http://"+loyaltyAddress+"/api/user/orders", orderID)
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

	item.Count -= count

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
	defer resp.Body.Close()

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

func (uc UseCase) regNewOrderAccrual(cookie, id, host, orderID string, count int) {
	item, err := uc.storage.GetItem(context.Background(), id)
	if err != nil {
		log.Println(err)
		return
	}

	item.Description = item.Name
	item.Price *= float32(count)

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

type UserMutesMap map[string]*sync.Mutex

var userMutexes = make(UserMutesMap)

func (uc UseCase) BulkBuy(ctx context.Context, cookie, username, accrualAddress, loyaltyAddress string, items []string, login bool) error {
	if userMutexes[username] == nil {
		userMutexes[username] = &sync.Mutex{}
	}

	userMutexes[username].Lock()
	defer userMutexes[username].Unlock()

	countOfSuccess := 0

	for _, item := range items {
		split := strings.Split(item, ":")
		if len(split) != 2 {
			log.Println(ErrBadOrder)
			continue
		}

		count, err := strconv.Atoi(split[0])
		if err != nil {
			log.Println("Atoi:", err)
			continue
		}

		err = uc.Buy(ctx, cookie, split[1], accrualAddress, loyaltyAddress, count, login)
		if err != nil {
			log.Println("Buy:", err)
			continue
		}

		countOfSuccess++
	}

	if countOfSuccess != len(items) {
		return ErrBadOrder
	}

	return nil
}
