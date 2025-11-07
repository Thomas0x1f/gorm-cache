package gormcache

import "context"

type contextKey string

const (
	contextKeySkipCache contextKey = "gorm:cache:skip"
)

// SkipCacheContext returns a new context that will skip cache for queries
func SkipCacheContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeySkipCache, true)
}

// WithSkipCache is a lower-level function to set skip cache value
func WithSkipCache(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, contextKeySkipCache, skip)
}

// getSkipCacheFromContext checks if cache should be skipped from context
func getSkipCacheFromContext(ctx context.Context) (bool, bool) {
	if ctx == nil {
		return false, false
	}
	if val := ctx.Value(contextKeySkipCache); val != nil {
		if skip, ok := val.(bool); ok {
			return skip, true
		}
	}
	return false, false
}
