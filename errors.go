package evbus

import "errors"

var (
	ErrEventNotRegisterd         = errors.New("EVENT_NOT_REGISTERED")
	ErrInvalidEventHandler       = errors.New("INVALID_EVENT_HANDLER")
	ErrInvalidEventHandlerArgs   = errors.New("INVALID_EVENT_HANDLER_ARGS")
	ErrEventDoesNotImplementTag  = errors.New("EVENT_DOES_NOT_IMPLEMENT_TAG")
	ErrInvalidHandlerTag         = errors.New("INVALID_HANDLER_TAG")
	ErrHandlerAlreadySubscribed  = errors.New("HANDLER_ALREADY_SUBSCRIBED")
	ErrHandlerAreadyUnsubscribed = errors.New("HANDLER_ALREADY_UNSUBSCRIBED")
	ErrUnknown                   = errors.New("UNKNOWN")
)
