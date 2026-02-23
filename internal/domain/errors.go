package domain

import "errors"

var (
	// ErrNotFound is returned when a requested entity does not exist.
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists is returned when attempting to create a duplicate entity.
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrUnauthorized is returned when authentication fails.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user lacks permission.
	ErrForbidden = errors.New("forbidden")

	// ErrValidation is returned when input validation fails.
	ErrValidation = errors.New("validation failed")

	// ErrInternal is returned for unexpected internal errors.
	ErrInternal = errors.New("internal server error")

	// ErrInvalidInput is returned when input data is malformed.
	ErrInvalidInput = errors.New("invalid input")

	// ErrEmailSendFailed is returned when sending an email fails.
	ErrEmailSendFailed = errors.New("email send failed")

	// ErrLLMProcessingFailed is returned when LLM processing fails.
	ErrLLMProcessingFailed = errors.New("llm processing failed")
)

// DomainError wraps a sentinel error with additional context.
type DomainError struct {
	Err     error
	Message string
}

// Error returns the error message.
func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap returns the underlying sentinel error.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new DomainError wrapping the given sentinel with a message.
func NewDomainError(sentinel error, message string) *DomainError {
	return &DomainError{
		Err:     sentinel,
		Message: message,
	}
}
