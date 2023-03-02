package usecase

import (
	"gomarket/internal/loyalty/storage"
)

type UseCase struct {
	storage storage.IStorage
}

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type IUseCase interface {
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	CheckID(host, cookie, id string) error
	GetBalance(cookie string) ([]byte, error)
	DrawBonuses(cookie string, sum float64, orderID string) error
	GetWithdrawals(cookie string) ([]byte, error)
	GetOrders(cookie string) ([]byte, error)
}

func New(storage storage.IStorage) UseCase {
	return UseCase{storage: storage}
}
