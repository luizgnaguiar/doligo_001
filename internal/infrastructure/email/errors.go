
package email

import "errors"

var (
	// ErrTransient is a transient error.
	ErrTransient = errors.New("transient error")
	// ErrDefinitive is a definitive error.
	ErrDefinitive = errors.New("definitive error")
)
