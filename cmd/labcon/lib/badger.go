package lib

import (
	"context"
	"net/http"

	"github.com/dgraph-io/badger/v3"
)

const BadgerContextKey AppContextKey = "badger"

func WithBadger(ctx context.Context, db *badger.DB) context.Context {
	return context.WithValue(ctx, BadgerContextKey, db)
}

func UseBadger(ctx context.Context) *badger.DB {
	return ctx.Value(BadgerContextKey).(*badger.DB)
}

func Badger(db *badger.DB) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			next.ServeHTTP(w, r.WithContext(WithBadger(ctx, db)))
		})
	}
}
