package storage

import "gomarket/internal/storage/service"

type IStorage interface {
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	CheckID(username, id string) error
	GetOrders(username string) (service.Orders, error)
}

type Type string
