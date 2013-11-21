package flags

import (
	"fmt"
)

// ErrorType represents the type of error.
type ErrorType uint

const (
	// Unknown or generic error
	ErrUnknown ErrorType = iota

	// Expected an argument but got none
	ErrExpectedArgument

	// Unknown flag
	ErrUnknownFlag

	// Unknown group
	ErrUnknownGroup

	// Failed to marshal value
	ErrMarshal

	// The error contains the builtin help message
	ErrHelp

	// An argument for a boolean value was specified
	ErrNoArgumentForBool

	// A required flag was not specified
	ErrRequired

	// A short flag name is longer than one character
	ErrShortNameTooLong
)

func (e ErrorType) String() string {
	switch e {
	case ErrUnknown:
		return "unknown"
	case ErrExpectedArgument:
		return "expected argument"
	case ErrUnknownFlag:
		return "unknown flag"
	case ErrUnknownGroup:
		return "unknown group"
	case ErrMarshal:
		return "marshal"
	case ErrHelp:
		return "help"
	case ErrNoArgumentForBool:
		return "no argument for bool"
	case ErrRequired:
		return "required"
	}

	return "unknown"
}

// Error represents a parser error. The error returned from Parse is of this
// type. The error contains both a Type and Message.
type Error struct {
	// The type of error
	Type ErrorType

	// The error message
	Message string
}

// Get the errors error message.
func (e *Error) Error() string {
	return e.Message
}

func newError(tp ErrorType, message string) *Error {
	return &Error{
		Type:    tp,
		Message: message,
	}
}

func newErrorf(tp ErrorType, format string, args ...interface{}) *Error {
	return newError(tp, fmt.Sprintf(format, args...))
}

func wrapError(err error) *Error {
	ret, ok := err.(*Error)

	if !ok {
		return newError(ErrUnknown, err.Error())
	}

	return ret
}
