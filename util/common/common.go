// Package common provides shared error and utility types.
package common

import (
	"errors"
	"fmt"
)

func NewError(a ...interface{}) error {
	return errors.New(fmt.Sprint(a...))
}

func NewErrorf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func Combine(errs ...error) error {
	var msg string
	for _, err := range errs {
		if err != nil {
			if msg != "" {
				msg += "; "
			}
			msg += err.Error()
		}
	}
	if msg == "" {
		return nil
	}
	return errors.New(msg)
}
