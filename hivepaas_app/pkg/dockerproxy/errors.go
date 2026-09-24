package dockerproxy

import "fmt"

// refusalError is a request the policy does not allow, as opposed to a failure
// to judge it. It is the app's to fix, and it is answered with 403.
type refusalError struct {
	msg string
}

func (e *refusalError) Error() string {
	return e.msg
}

// refusef returns a refusal saying which rule the request broke.
func refusef(format string, args ...any) error {
	return &refusalError{msg: fmt.Sprintf(format, args...)}
}
