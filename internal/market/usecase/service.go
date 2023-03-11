package usecase

import (
	"context"
	"errors"
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
	Authentication(login, passwd string) (string, error)
	GetBalance(ctx context.Context, cookie string, loyaltyAddress string) (schema.BalanceMarket, error)
	GetItems(ctx context.Context) ([]schema.Item, error)
	Buy(ctx context.Context, cookie, id, accrualAddress, loyaltyAddress string, count int, login bool) (schema.Item, error)
	BulkBuy(ctx context.Context, cookie, username, accrualAddress, loyaltyAddress string, items []string, login bool) error
	GetOrders(ctx context.Context, username string) ([]schema.Order, error)
	GetAllOrders(ctx context.Context) ([]schema.Order, error)
	AddItem(ctx context.Context, item schema.Item) error
	RemoveItem(ctx context.Context, id string) error
	ChangeItem(ctx context.Context, item schema.Item) error
	IsAdmin(ctx context.Context, username string) (bool, error)
	ChangeOrderStatus(ctx context.Context, status, orderID string) error
}

func New(storage storage.IStorage) UseCase {
	return UseCase{storage: storage}
}

var ErrBadOrder = errors.New("some items were not purchased")
var ErrReservedUsername = errors.New("username is reserved")
var ErrServer = errors.New("server error, sorry! we're already working on it")
var ErrBadCookie = errors.New("bad cookie")
var ErrDeadLoyalty = errors.New("we are sorry, registration is not available at the moment")
