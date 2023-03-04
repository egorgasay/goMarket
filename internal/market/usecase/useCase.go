package usecase

import (
	"context"
	"gomarket/internal/loyalty/storage"
	"gomarket/internal/market/schema"
	"strings"
)

func (uc UseCase) CreateUser(login, passwd, cookie, loyaltyCookie string) (string, error) {
	split := strings.Split(loyaltyCookie, "-")
	if len(split) != 2 {
		return "", storage.ErrBadCookie
	}

	return uc.storage.CreateUser(login, passwd, cookie, split[1])
}

func (uc UseCase) CheckPassword(login, passwd string) error {
	return uc.storage.CheckPassword(login, passwd)
}

func (uc UseCase) GetBalance(ctx context.Context, cookie string) (schema.BalanceMarket, error) {
	balance, err := uc.storage.GetBalance(ctx, cookie)
	balance.Bonuses = 0 // TODO: Connect with loyalty service
	return balance, err
}

func (uc UseCase) CreateAnonUser(ctx context.Context, cookie string) error {
	return uc.storage.CreateAnonUser(ctx, schema.Customer{Cookie: cookie, Current: 10000})
}

func (uc UseCase) GetItems(ctx context.Context) ([]schema.Item, error) {
	return uc.storage.GetItems(ctx)
}

func (uc UseCase) Buy(ctx context.Context, cookie string, id string) error {
	// go RegNewAccural
	// go RegNewLoyalty

	return uc.storage.Buy(ctx, cookie, id)
}

func regNewOrder(ctx context.Context, cookie string, id string) error {
	return nil
}
