package usecase

import (
	"gomarket/internal/storage"
)

type UseCase struct {
	storage storage.IStorage
}

func New(storage storage.IStorage) UseCase {
	return UseCase{storage: storage}
}
