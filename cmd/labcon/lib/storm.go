package lib

import (
	"context"
	"net/http"

	"github.com/asdine/storm/v3"
)

const StormContextKey AppContextKey = "storm"

func WithStorm(ctx context.Context, db *storm.DB) context.Context {
	return context.WithValue(ctx, StormContextKey, db)
}

func UseStorm(ctx context.Context) *storm.DB {
	return ctx.Value(StormContextKey).(*storm.DB)
}

func Storm(path string, opts ...func(*storm.Options) error) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			db, err := storm.Open(path, opts...)
			if err != nil {
				logger := UseLogger(ctx)
				logger.Error().Err(err).Msg("failed to open storm database")
			}
			defer db.Close()
			next.ServeHTTP(w, r.WithContext(WithStorm(ctx, db)))
		})
	}
}

func ConvertStormError(err error) error {
	switch err {
	case storm.ErrAlreadyExists:
		return ErrAlreadyExists
	case storm.ErrNotFound:
		return ErrNotFound
	default:
		return err
	}
}
