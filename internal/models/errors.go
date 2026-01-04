package models

import "errors"

var (
	ErrInvalidStatus     = errors.New("invalid order status")
	ErrInvalidTransition = errors.New("invalid order transition")
	ErrTerminalState     = errors.New("order is in terminal state")
)
