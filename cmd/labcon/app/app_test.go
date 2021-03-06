package app_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/ktnyt/labcon/cmd/labcon/app"
	"github.com/ktnyt/labcon/cmd/labcon/app/injectors"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
	"github.com/ktnyt/labcon/driver"
	"github.com/ktnyt/labcon/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestIntegration(t *testing.T) {
	r := chi.NewMux()

	b := &strings.Builder{}
	logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
	logger := log.Output(logout).Level(zerolog.TraceLevel)

	token := lib.DefaultTokenGenerator()

	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	r.Use(
		lib.Logger(logger),
		lib.Badger(db),
		lib.DriverTokenGenerator(func() string { return token }),
	)

	a := app.NewApp(injectors.Driver)
	a.Setup(r)

	server := httptest.NewServer(r)
	defer server.Close()

	newRequest := func(t *testing.T, method, path string, body io.Reader) *http.Request {
		url := fmt.Sprintf("%s%s", server.URL, path)
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			t.Fatal(err)
		}
		return req
	}

	client := http.Client{}

	cases := []struct {
		setup func() *http.Request
		code  int
		out   io.Reader
	}{
		{
			setup: func() *http.Request {
				return newRequest(t, http.MethodGet, "/", nil)
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString("OK\n"),
		},

		{
			setup: func() *http.Request {
				req := newRequest(t, http.MethodPost, "/driver", lib.MustJsonMarshalToBuffer(t, driver.RegisterParams{
					Name:  "foo",
					State: "foo",
				}))
				req.Header.Add("Content-Type", "application/json")
				return req
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString(fmt.Sprintf("%q\n", token)),
		},

		{
			setup: func() *http.Request {
				return newRequest(t, http.MethodGet, "/driver", nil)
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, []string{"foo"}),
		},

		{
			setup: func() *http.Request {
				return newRequest(t, http.MethodGet, "/driver/foo/state", nil)
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, "foo"),
		},

		{
			setup: func() *http.Request {
				return newRequest(t, http.MethodGet, "/driver/foo/status", nil)
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, driver.Idle),
		},

		{
			setup: func() *http.Request {
				req := newRequest(t, http.MethodGet, "/driver/foo/operation", nil)
				req.Header.Add("X-Driver-Token", token)
				return req
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, nil),
		},

		{
			setup: func() *http.Request {
				req := newRequest(t, http.MethodPost, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Op{
					Name: "op",
					Arg:  "arg",
				}))
				req.Header.Add("X-Driver-Token", token)
				req.Header.Add("Content-Type", "application/json")
				return req
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString("OK\n"),
		},

		{
			setup: func() *http.Request {
				return newRequest(t, http.MethodGet, "/driver/foo/status", nil)
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, driver.Busy),
		},

		{
			setup: func() *http.Request {
				req := newRequest(t, http.MethodGet, "/driver/foo/operation", nil)
				req.Header.Add("X-Driver-Token", token)
				return req
			},
			code: http.StatusOK,
			out: lib.MustJsonMarshalToBuffer(t, driver.Op{
				Name: "op",
				Arg:  "arg",
			}),
		},

		{
			setup: func() *http.Request {
				req := newRequest(t, http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				req.Header.Add("X-Driver-Token", token)
				req.Header.Add("Content-Type", "application/json")
				return req
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString("OK\n"),
		},

		{
			setup: func() *http.Request {
				return newRequest(t, http.MethodGet, "/driver/foo/state", nil)
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, "bar"),
		},

		{
			setup: func() *http.Request {
				req := newRequest(t, http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				req.Header.Add("X-Driver-Token", token)
				req.Header.Add("Content-Type", "application/json")
				return req
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString("OK\n"),
		},

		{
			setup: func() *http.Request {
				return newRequest(t, http.MethodGet, "/driver/foo/status", nil)
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, driver.Idle),
		},

		{
			setup: func() *http.Request {
				req := newRequest(t, http.MethodGet, "/driver/foo/operation", nil)
				req.Header.Add("X-Driver-Token", token)
				return req
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, nil),
		},

		{
			setup: func() *http.Request {
				req := newRequest(t, http.MethodDelete, "/driver/foo", nil)
				req.Header.Add("X-Driver-Token", token)
				return req
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString("OK\n"),
		},
	}

	failed := false

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			req := tt.setup()
			res, err := client.Do(req)
			if err != nil {
				t.Errorf("client.Do(%v): %v", req, err)
				failed = true
			}

			if res.StatusCode != tt.code {
				t.Errorf("client.Do(%v).StatusCode = %s, expecting %s", req, http.StatusText(res.StatusCode), http.StatusText(tt.code))
				failed = true
			}

			if ops := utils.ReaderDiff(res.Body, tt.out); ops != nil {
				t.Error(utils.JoinOps(ops, "\n"))
				failed = true
			}
		})

		if failed {
			return
		}
	}
}
