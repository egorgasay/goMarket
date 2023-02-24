package storage

import (
	"gomarket/internal/schema"
	"gomarket/internal/storage/service"
)

type IStorage interface {
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	CheckID(username, id string) error
	GetOrders(username string) (service.Orders, error)
	GetBalance(username string) (schema.Balance, error)
	UpdateOrder(id, status string, accrual float64) error
}

type Type string
