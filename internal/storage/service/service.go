package service

import "errors"

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type IRealStorage interface {
	CreateUser(login, passwd string) error
	CheckPassword(login, passwd string) error
	CheckID(username, id string) error
}

var ErrUsernameConflict = errors.New("username already exists")
var ErrWrongPassword = errors.New("wrong password")
var ErrBadCookie = errors.New("bad cookie")
var ErrCreatedByAnotherUser = errors.New("uid already exists and created by another user")
var ErrCreatedByThisUser = errors.New("uid already exists and created by this user")
var ErrBadID = errors.New("wrong id format")
