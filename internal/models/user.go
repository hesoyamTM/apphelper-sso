package models

import "github.com/google/uuid"

type UserInfo struct {
	Id      uuid.UUID
	Name    string
	Surname string
}

type UserAuth struct {
	Id       uuid.UUID
	Login    string
	PassHash []byte
}

type User struct {
	UserInfo
	UserAuth
}
