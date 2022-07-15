package publisher

import "errors"

var (
	ErrInternal = errors.New("internal error while sending notification")
)
