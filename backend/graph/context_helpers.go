package graph

import "context"

type userIDContextKey string

const userIDKey userIDContextKey = "userID"

// WithUserID stores user ID in context for resolvers.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts user ID from context when present.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok && id != ""
}
