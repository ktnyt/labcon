package lib

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := logger.With()
			c = c.Str("protocol", r.Proto)
			c = c.Str("method", r.Method)
			c = c.Str("url", r.URL.String())

			logger = c.Logger()

			ctx := logger.WithContext(r.Context())

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t := time.Now()
			defer func() {
				logger.Info().Dur("elapsed", time.Since(t)).Msg(http.StatusText(ww.Status()))
			}()

			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}

func UseLogger(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

type Adaptor zerolog.Logger

func (adaptor Adaptor) Errorf(format string, args ...interface{}) {
	logger := zerolog.Logger(adaptor)
	logger.Error().Msgf(format, args)
}

func (adaptor Adaptor) Warningf(format string, args ...interface{}) {
	logger := zerolog.Logger(adaptor)
	logger.Warn().Msgf(format, args)
}

func (adaptor Adaptor) Infof(format string, args ...interface{}) {
	logger := zerolog.Logger(adaptor)
	logger.Info().Msgf(format, args)
}

func (adaptor Adaptor) Debugf(format string, args ...interface{}) {
	logger := zerolog.Logger(adaptor)
	logger.Debug().Msgf(format, args)
}
