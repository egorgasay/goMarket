package usecase

import (
	"context"
	"gomarket/internal/market/schema"
	"gomarket/internal/market/storage"
)

type UseCase struct {
	storage storage.IStorage
}

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type IUseCase interface {
	CreateAnonUser(ctx context.Context, cookie string) error
	CreateUser(user schema.Customer, cookie, loyaltyAddress string) (string, error)
	CheckPassword(login, passwd string) error
	GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error)
	GetItems(ctx context.Context) ([]schema.Item, error)
	Buy(ctx context.Context, cookie string, id string) error
}

func New(storage storage.IStorage) UseCase {
	return UseCase{storage: storage}
}
