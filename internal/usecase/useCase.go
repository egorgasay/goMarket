package usecase

import (
	"encoding/hex"
	"encoding/json"
	"gomarket/internal/storage/service"
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

	return uc.storage.CheckID(username, id)
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
