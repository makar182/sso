package storage

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrAppNotFound       = errors.New("app not found")
	//ErrSomeStorageProblem = errors.New("some storage problem")
)
