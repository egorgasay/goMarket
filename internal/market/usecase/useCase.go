package usecase

import (
	"context"
	"gomarket/internal/schema"
)

func (uc UseCase) CreateUser(login, passwd string) error {
	return uc.storage.CreateUser(login, passwd)
}

func (uc UseCase) CheckPassword(login, passwd string) error {
	return uc.storage.CheckPassword(login, passwd)
}

func (uc UseCase) GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error) {
	return uc.storage.GetBalance(ctx, cookie)
}

func (uc UseCase) CreateAnonUser(ctx context.Context, cookie string) error {
	return uc.storage.CreateAnonUser(ctx, schema.Customer{Cookie: cookie, Current: 10000})
}
