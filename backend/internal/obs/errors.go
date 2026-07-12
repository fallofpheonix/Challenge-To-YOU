package obs

import "fmt"

// ErrorClass categorizes errors by subsystem for structured logging.
type ErrorClass string

const (
	ClassInfra       ErrorClass = "infrastructure"
	ClassTransport   ErrorClass = "transport"
	ClassGameplay    ErrorClass = "gameplay"
	ClassCompiler    ErrorClass = "compiler"
	ClassSandbox     ErrorClass = "sandbox"
	ClassPersistence ErrorClass = "persistence"
	ClassAI          ErrorClass = "ai"
	ClassUnknown     ErrorClass = "unknown"
)

// ClassifiedError is an error annotated with an ErrorClass.
type ClassifiedError struct {
	Class ErrorClass
	Err   error
}

func (e *ClassifiedError) Error() string {
	return e.Err.Error()
}

func (e *ClassifiedError) Unwrap() error {
	return e.Err
}

// Classify wraps an error with a class. If err is nil, returns nil.
func Classify(class ErrorClass, err error) *ClassifiedError {
	if err == nil {
		return nil
	}
	return &ClassifiedError{Class: class, Err: err}
}

// Classifyf creates a classified error from a format string.
func Classifyf(class ErrorClass, format string, args ...any) *ClassifiedError {
	return &ClassifiedError{Class: class, Err: fmt.Errorf(format, args...)}
}
