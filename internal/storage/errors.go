package storage

import "errors"

var (
	ErrUserExists                  = errors.New("user already exists")
	ErrUserNotFound                = errors.New("user not found")
	ErrSessionNotFound             = errors.New("session not found")
	ErrVerificationCodeNotFound    = errors.New("verification code not found")
	ErrChangePasswordTokenNotFound = errors.New("change password token not found")
)
