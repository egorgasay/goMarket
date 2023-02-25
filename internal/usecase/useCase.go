package usecase

import (
	"encoding/hex"
	"encoding/json"
	"gomarket/internal/schema"
	"gomarket/internal/storage"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (uc UseCase) CreateUser(login, passwd string) error {
	return uc.storage.CreateUser(login, passwd)
}

func (uc UseCase) CheckPassword(login, passwd string) error {
	return uc.storage.CheckPassword(login, passwd)
}

func (uc UseCase) CheckID(host, cookie, id string) error {
	if !allCharsIsDigits(id) {
		return storage.ErrBadID
	}

	username, err := getUsernameFromCookie(cookie)
	if err != nil {
		return err
	}

	ID, err := strconv.Atoi(id)
	if err != nil {
		return storage.ErrBadID
	}

	if !Valid(ID) {
		return storage.ErrBadID
	}
	err = uc.storage.CheckID(username, id)
	if err != nil {
		return err
	}

	go uc.updateStatus(username, host, id)

	return nil
}

func (uc UseCase) updateStatus(username, host, id string) {
	ticker := time.NewTicker(1 * time.Second)
	status := ""
	for status != "PROCESSED" && status != "INVALID" {
		<-ticker.C
		res, err := http.Get(host + "/api/orders/" + id)
		if err != nil {
			log.Println(err)
			continue
		}
		switch res.StatusCode {
		case http.StatusNoContent:
			log.Println("No content")
			continue
		case http.StatusInternalServerError:
			log.Println("Calc storage error")
			continue
		case http.StatusTooManyRequests:
			log.Println("Too many request")
			continue
		}

		read, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		res.Body.Close()

		var response schema.ResponseFromTheCalculationSystem
		err = json.Unmarshal(read, &response)
		if err != nil {
			log.Println(err)
			continue
		}

		err = uc.storage.UpdateOrder(username, id, response.Status, response.Accrual)
		if err != nil {
			log.Println(err)
		}

		status = response.Status
	}
}

func (uc UseCase) GetBalance(cookie string) ([]byte, error) {
	username, err := getUsernameFromCookie(cookie)
	if err != nil {
		return []byte(""), err
	}

	balance, err := uc.storage.GetBalance(username)
	if err != nil {
		return []byte(""), err
	}

	res, err := json.Marshal(balance)
	if err != nil {
		return []byte(""), err
	}

	return res, nil
}

func (uc UseCase) DrawBonuses(cookie string, sum float64, orderID string) error {
	if !allCharsIsDigits(orderID) {
		return storage.ErrBadID
	}

	username, err := getUsernameFromCookie(cookie)
	if err != nil {
		return err
	}

	ID, err := strconv.Atoi(orderID)
	if err != nil {
		return storage.ErrBadID
	}

	if !Valid(ID) {
		return storage.ErrBadID
	}

	return uc.storage.Withdraw(username, sum, orderID)
}

func (uc UseCase) GetWithdrawals(cookie string) ([]byte, error) {
	username, err := getUsernameFromCookie(cookie)
	if err != nil {
		return []byte(""), err
	}

	withdrawals, err := uc.storage.GetWithdrawals(username)
	if err != nil {
		return []byte(""), err
	}

	res, err := json.Marshal(withdrawals)
	return res, err
}

func (uc UseCase) GetOrders(cookie string) ([]byte, error) {
	username, err := getUsernameFromCookie(cookie)
	if err != nil {
		return []byte(""), err
	}

	orders, err := uc.storage.GetOrders(username)
	if err != nil {
		return []byte(""), err
	}

	res, err := json.Marshal(orders)
	if err != nil {
		return []byte(""), err
	}

	return res, nil
}

func Valid(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}

func getUsernameFromCookie(cookie string) (string, error) {
	split := strings.Split(cookie, "-")
	if len(split) != 2 {
		return "", storage.ErrBadCookie
	}

	username := split[1]
	user, err := hex.DecodeString(username)
	return string(user), err
}

func allCharsIsDigits(input string) bool {
	for _, sym := range input {
		if !strings.ContainsAny(string(sym), "0123456789") {
			return false
		}
	}
	return true
}
