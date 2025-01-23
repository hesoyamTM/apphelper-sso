package models

type UserInfo struct {
	Id      int64
	Name    string
	Surname string
}

type UserAuth struct {
	Id       int64
	Login    string
	PassHash []byte
}

type User struct {
	UserInfo
	UserAuth
}
