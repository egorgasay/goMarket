package storage

type IStorage interface {
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	CheckID(username, id string) error
}

type Type string
