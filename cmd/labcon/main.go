package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ktnyt/labcon/cmd/labcon/app"
	"github.com/ktnyt/labcon/cmd/labcon/app/injectors"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	w := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	logger := log.Output(w).Level(zerolog.TraceLevel)

	r := chi.NewMux()

	corsOpts := cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"X-PINGOTHER",
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Driver-Token",
		},
		AllowCredentials: true,
	}

	opts := badger.DefaultOptions("labcon.db").WithLogger(lib.Adaptor(logger))
	db, err := badger.Open(opts)
	if err != nil {
		logger.Err(err).Msg("failed to open database")
	}
	defer db.Close()

	r.Use(
		lib.Logger(logger),
		cors.Handler(corsOpts),
		lib.Badger(db),
		lib.DriverTokenGenerator(lib.DefaultTokenGenerator),
		lib.CurrentTime,
		middleware.Timeout(time.Second*60),
		middleware.Recoverer,
	)

	a := app.NewApp(injectors.Driver)
	a.Setup(r)

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	addr := fmt.Sprintf("%s:%s", host, port)

	http.ListenAndServe(addr, r)
}
