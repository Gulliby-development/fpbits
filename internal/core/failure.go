package core

import "fmt"

// Failure is a user-facing failure with a stable diagnostic message.
type Failure struct {
	Message string
}

// Error returns the failure message.
func (f Failure) Error() string {
	return f.Message
}

// Failf formats a message and returns it as a Failure.
func Failf(format string, args ...any) error {
	return Failure{Message: fmt.Sprintf(format, args...)}
}
