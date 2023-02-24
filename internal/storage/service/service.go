package service

import (
	"errors"
	"gomarket/internal/schema"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type IRealStorage interface {
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	CheckID(username, id string) error
	GetOrders(username string) (Orders, error)
	GetBalance(username string) (schema.Balance, error)
	UpdateOrder(id, status string, accrual float64) error
}

type Orders []*schema.UserOrder

var ErrUsernameConflict = errors.New("username already exists")
var ErrWrongPassword = errors.New("wrong password")
var ErrBadCookie = errors.New("bad cookie")
var ErrCreatedByAnotherUser = errors.New("uid already exists and created by another user")
var ErrCreatedByThisUser = errors.New("uid already exists and created by this user")
var ErrBadID = errors.New("wrong id format")
var ErrNoResult = errors.New("the user has no orders")
