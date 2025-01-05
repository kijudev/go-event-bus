package evbus

import "errors"

var (
	ErrEventNotRegisterd       = errors.New("EVENT_NOT_REGISTERED")
	ErrInvalidEventHandler     = errors.New("INVALID_EVENT_HANDLER")
	ErrInvalidEventHandlerArgs = errors.New("INVALID_EVENT_HANDLER_ARGS")
)
