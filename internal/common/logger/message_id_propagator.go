package logger

import "context"

type contextKey string

const messageIDKey contextKey = "message_id"

func WithMessageID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, messageIDKey, id)
}

func GetMessageID(ctx context.Context) string {
	if v, ok := ctx.Value(messageIDKey).(string); ok {
		return v
	}
	return ""
}
