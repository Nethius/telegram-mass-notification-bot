package repository

import "errors"

var (
	ErrAlreadyExists = errors.New("user already exists")
	ErrNotExists     = errors.New("user does not exist")
	ErrInternal      = errors.New("something went wrong")
)
