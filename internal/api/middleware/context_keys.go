package middleware

import "context"

type contextKey string

const CorrelationIDKey contextKey = "correlationID"

// ToContext adds the correlation ID to the given context.
func ToContext(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// FromContext extracts the correlation ID from the context, if present.
func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(CorrelationIDKey).(string)
	return id, ok
}