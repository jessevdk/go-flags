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

	// The builtin help message was printed
	ErrHelpShown

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

func newError(tp ErrorType, message string) error {
	return &Error{
		Type:    tp,
		Message: message,
	}
}

// The builtin help message was printed. This value will be returned from
// Parse when the builtin help message was shown (i.e. when the ShowHelp
// option is set and either -h or --help was specified on the command line).
var ErrHelp = newError(ErrHelpShown, "help shown")
