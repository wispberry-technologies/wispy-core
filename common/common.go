package common

import "fmt"

func NewError(message string, args ...interface{}) error {
	if len(args) > 0 {
		return fmt.Errorf(message, args...)
	}
	return fmt.Errorf("%s", message)
}
func NewErrorf(format string, args ...interface{}) error {
	if len(args) > 0 {
		return fmt.Errorf(format, args...)
	}
	return fmt.Errorf("%s", format)
}
