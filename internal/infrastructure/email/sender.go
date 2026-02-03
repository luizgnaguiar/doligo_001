
package email

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State represents the state of the circuit breaker.
type State int

const (
	// Closed is the initial state of the circuit breaker.
	Closed State = iota
	// Open is the state when the circuit breaker is active.
	Open
)

const (
	// consecutiveFailuresThreshold is the number of consecutive failures
	// before the circuit breaker is opened.
	consecutiveFailuresThreshold = 3
	// openStateTimeout is the duration the circuit breaker will remain
	// in the open state.
	openStateTimeout = 60 * time.Second
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// EmailSender defines the interface for sending emails.
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// simpleEmailSender is a simple implementation of EmailSender.
type simpleEmailSender struct {
	mu                  sync.Mutex
	state               State
	consecutiveFailures int
	lastFailureTime     time.Time
}

// NewSimpleEmailSender creates a new simpleEmailSender.
func NewSimpleEmailSender() EmailSender {
	return &simpleEmailSender{
		state: Closed,
	}
}

// Send sends an email with retry and backoff for transient errors.
func (s *simpleEmailSender) Send(ctx context.Context, to, subject, body string) error {
	s.mu.Lock()
	if s.state == Open {
		if time.Since(s.lastFailureTime) > openStateTimeout {
			s.state = Closed
			s.consecutiveFailures = 0
		} else {
			s.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	s.mu.Unlock()

	const maxRetries = 3
	const backoff = 2 * time.Second

	var err error
	for i := 0; i < maxRetries; i++ {
		err = s.sendEmail(ctx, to, subject, body)
		if err == nil {
			s.mu.Lock()
			s.consecutiveFailures = 0
			s.mu.Unlock()
			return nil
		}

		if errors.Is(err, ErrDefinitive) {
			return err
		}

		if errors.Is(err, ErrTransient) {
			s.mu.Lock()
			s.consecutiveFailures++
			if s.consecutiveFailures >= consecutiveFailuresThreshold {
				s.state = Open
				s.lastFailureTime = time.Now()
				s.mu.Unlock()
				return err
			}
			s.mu.Unlock()
		}

		time.Sleep(backoff * time.Duration(i+1))
	}

	return err
}

// sendEmail is a placeholder for the actual email sending logic.
func (s *simpleEmailSender) sendEmail(ctx context.Context, to, subject, body string) error {
	// In a real implementation, this would use an SMTP client to send the email.
	// For now, we will simulate errors.
	if time.Now().Unix()%3 == 0 {
		return ErrTransient
	}
	if time.Now().Unix()%5 == 0 {
		return ErrDefinitive
	}
	return nil
}
