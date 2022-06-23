package command_parser

import (
	"errors"
)

var (
	ErrAlreadyRegistered = errors.New("user already registered")
	ErrNotExist          = errors.New("user does not exist")
)
