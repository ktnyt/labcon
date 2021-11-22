package lib

import (
	"context"
	"net/http"
)

const DriverTokenGeneratorContextKey AppContextKey = "driver_token_generator"

type TokenGenerator func() string

func DefaultTokenGenerator() string {
	return Base32String(NewToken(20))
}

func WithDriverTokenGenerator(ctx context.Context, gen TokenGenerator) context.Context {
	return context.WithValue(ctx, DriverTokenGeneratorContextKey, gen)
}

func UseDriverTokenGenerator(ctx context.Context) TokenGenerator {
	return ctx.Value(DriverTokenGeneratorContextKey).(TokenGenerator)
}

func DriverTokenGenerator(gen TokenGenerator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithDriverTokenGenerator(r.Context(), gen)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
