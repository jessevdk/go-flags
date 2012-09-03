package flags

type ErrorType uint

const (
	ErrUnknown ErrorType = iota

	// Expected an argument but got none
	ErrExpectedArgument

	// Unknown flag
	ErrUnknownFlag

	// Failed to marshal value
	ErrMarshal

	// The builtin help message was printed
	ErrHelpShown

	// An argument for a boolean value was specified
	ErrNoArgumentForBool
)

type Error struct {
	Type ErrorType
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func newError(tp ErrorType, message string) error {
	return &Error {
		Type: tp,
		Message: message,
	}
}

var ErrHelp = newError(ErrHelpShown, "help shown")
