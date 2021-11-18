package lib

import (
	"context"
	"net/http"
	"time"
)

var TimeContextKey AppContextKey = "time"

func WithTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, TimeContextKey, t)
}

func UseTime(ctx context.Context) time.Time {
	value, ok := ctx.Value(TimeContextKey).(time.Time)
	if !ok {
		value = time.Now()
	}
	return value
}

func CurrentTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithTime(r.Context(), time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
