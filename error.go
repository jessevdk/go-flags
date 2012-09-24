package flags

// ErrorType represents the type of error.
type ErrorType uint

const (
	// Unknown or generic error
	ErrUnknown ErrorType = iota

	// Expected an argument but got none
	ErrExpectedArgument

	// Unknown flag
	ErrUnknownFlag

	// Failed to marshal value
	ErrMarshal

	// The error contains the builtin help message
	ErrHelp

	// An argument for a boolean value was specified
	ErrNoArgumentForBool
)

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
