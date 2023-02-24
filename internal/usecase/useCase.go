package usecase

import (
	"encoding/hex"
	"encoding/json"
	"gomarket/internal/storage/service"
	"strconv"
	"strings"
)

func (uc UseCase) CreateUser(login, passwd string) error {
	return uc.storage.CreateUser(login, passwd)
}

func (uc UseCase) CheckPassword(login, passwd string) error {
	return uc.storage.CheckPassword(login, passwd)
}

func (uc UseCase) CheckID(cookie, id string) error {
	if !allCharsIsDigits(id) {
		return service.ErrBadID
	}

	username, err := getUsernameFromCookie(cookie)
	if err != nil {
		return err
	}

	ID, err := strconv.Atoi(id)
	if err != nil {
		return service.ErrBadID
	}

	if !Valid(ID) {
		return service.ErrBadID
	}

	return uc.storage.CheckID(username, id)
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
		return "", service.ErrBadCookie
	}

	username := split[1]
	user, err := hex.DecodeString(username)
	return string(user), err
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

func allCharsIsDigits(input string) bool {
	for _, sym := range input {
		if strings.ContainsAny(string(sym), "0123456789") == false {
			return false
		}
	}
	return true
}
