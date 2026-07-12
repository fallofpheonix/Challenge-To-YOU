package obs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type contextKey int

const (
	requestIDKey contextKey = iota
	sessionIDKey
)

// RequestIDFrom extracts the request ID from the context. Returns "" if absent.
func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// SessionIDFrom extracts the session ID from the context. Returns "" if absent.
func SessionIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(sessionIDKey).(string); ok {
		return v
	}
	return ""
}

// WithRequestID returns a new context carrying the given request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// WithSessionID returns a new context carrying the given session ID.
func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

// NewRequestID generates a random 16-byte hex request ID.
func NewRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
