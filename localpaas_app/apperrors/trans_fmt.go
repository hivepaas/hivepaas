package apperrors

import "fmt"

func Fmt(format string, args ...any) any {
	// TODO: add implementation to handle translation of string format
	// For now, just call fmt.Sprintf to return the final string
	return fmt.Sprintf(format, args...)
}
