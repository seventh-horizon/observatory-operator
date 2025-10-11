package controllers

import "fmt"

type UserError struct {
	Operation string
	Cause     error
	Message   string
	Hints     []string
}

func (e *UserError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Operation, e.Message)
	if e.Cause != nil { msg += fmt.Sprintf(" (cause: %v)", e.Cause) }
	for i, h := range e.Hints {
		msg += fmt.Sprintf("\n  %d. %s", i+1, h)
	}
	return msg
}
