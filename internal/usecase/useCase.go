package usecase

import (
	"encoding/hex"
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

	split := strings.Split(cookie, "-")
	if len(split) != 2 {
		return service.ErrBadCookie
	}

	username := split[1]
	user, err := hex.DecodeString(username)
	if err != nil {
		return err
	}

	return uc.storage.CheckID(string(user), id)
}

func allCharsIsDigits(input string) bool {
	for _, sym := range input {
		if strings.ContainsAny(string(sym), "0123456789") == false {
			return false
		}
	}
	return true
}
