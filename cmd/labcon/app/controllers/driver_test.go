package controllers_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/ktnyt/labcon/cmd/labcon/app/controllers"
	"github.com/ktnyt/labcon/cmd/labcon/app/usecases"
	"github.com/ktnyt/labcon/cmd/labcon/app/usecases_mock"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
	"github.com/ktnyt/labcon/driver"
	"github.com/ktnyt/labcon/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestDriverRegister(t *testing.T) {
	token := lib.Base32String(lib.NewToken(20))

	cases := []struct {
		label string
		mock  func(usecase *usecases_mock.MockDriverUsecase)
		setup func() *http.Request
		code  int
		out   io.Reader
	}{
		{
			label: "success",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Register("foo", "foo").
					Return(token, nil).
					Times(1)
			},
			setup: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/driver", lib.MustJsonMarshalToBuffer(
					t, driver.RegisterParams{
						Name:  "foo",
						State: "foo",
					},
				))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, token),
		},

		{
			label: "missing Content-Type header",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/driver", lib.MustJsonMarshalToBuffer(
					t, driver.RegisterParams{
						Name:  "foo",
						State: "foo",
					},
				))
				return r
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("Bad Request\n"),
		},

		{
			label: "validation error",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/driver", lib.MustJsonMarshalToBuffer(
					t, driver.RegisterParams{
						Name:  "",
						State: nil,
					},
				))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			code: http.StatusBadRequest,
			out: bytes.NewBufferString(strings.Join([]string{
				"validation failed on field \"name\" for constraint \"required\"",
				"validation failed on field \"state\" for constraint \"required\"",
				"",
			}, "\n")),
		},

		{
			label: "already exists",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Register("foo", "foo").
					Return("", lib.ErrAlreadyExists).
					Times(1)
			},
			setup: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/driver", lib.MustJsonMarshalToBuffer(
					t, driver.RegisterParams{
						Name:  "foo",
						State: "foo",
					},
				))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("failed to register driver \"foo\": already exists\n"),
		},

		{
			label: "internal error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Register("foo", "foo").
					Return("", lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/driver", lib.MustJsonMarshalToBuffer(
					t, driver.RegisterParams{
						Name:  "foo",
						State: "foo",
					},
				))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},
	}

	for _, tt := range cases {
		lib.RunCase(t, tt.label, func(t *testing.T) {
			failed := false

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := usecases_mock.NewMockDriverUsecase(ctrl)
			inject := func(context.Context) usecases.DriverUsecase { return usecase }
			controller := controllers.NewDriverController(inject)

			tt.mock(usecase)

			w := httptest.NewRecorder()
			r := tt.setup()

			b := &strings.Builder{}
			logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
			logger := log.Output(logout).Level(zerolog.TraceLevel)

			ctx := r.Context()
			ctx = logger.WithContext(ctx)

			controller.Register(w, r.WithContext(ctx))

			if w.Code != tt.code {
				t.Errorf("%s %s got %d: expected %d", r.Method, r.RequestURI, w.Code, tt.code)
				failed = true
			}

			if ops := utils.ReaderDiff(w.Body, tt.out); ops != nil {
				t.Errorf("%s %s response body:\n%s", r.Method, r.RequestURI, utils.JoinOps(ops, "\n"))
				failed = true
			}

			if failed {
				t.Errorf("log output:\n%s", b.String())
			}
		})
	}
}

func TestDriverGetState(t *testing.T) {
	cases := []struct {
		label string
		mock  func(usecase *usecases_mock.MockDriverUsecase)
		setup func() *http.Request
		code  int
		out   io.Reader
	}{
		{
			label: "success",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					GetState("foo").
					Return("foo", nil).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/state", nil)
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, "foo"),
		},

		{
			label: "missing URL parameter",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/state", nil)
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing URL parameter \"name\"\n"),
		},

		{
			label: "not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					GetState("foo").
					Return(nil, lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/state", nil)
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to get state for driver \"foo\": not found\n"),
		},

		{
			label: "internal error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					GetState("foo").
					Return(nil, lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/state", nil)
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},
	}

	for _, tt := range cases {
		lib.RunCase(t, tt.label, func(t *testing.T) {
			failed := false

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := usecases_mock.NewMockDriverUsecase(ctrl)
			inject := func(context.Context) usecases.DriverUsecase { return usecase }
			controller := controllers.NewDriverController(inject)

			tt.mock(usecase)

			w := httptest.NewRecorder()
			r := tt.setup()

			b := &strings.Builder{}
			logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
			logger := log.Output(logout).Level(zerolog.TraceLevel)

			ctx := r.Context()
			ctx = logger.WithContext(ctx)

			controller.GetState(w, r.WithContext(ctx))

			if w.Code != tt.code {
				t.Errorf("%s %s got %d: expected %d", r.Method, r.RequestURI, w.Code, tt.code)
				failed = true
			}

			if ops := utils.ReaderDiff(w.Body, tt.out); ops != nil {
				t.Errorf("%s %s response body:\n%s", r.Method, r.RequestURI, utils.JoinOps(ops, "\n"))
				failed = true
			}

			if failed {
				t.Errorf("log output:\n%s", b.String())
			}
		})
	}
}

func TestDriverSetState(t *testing.T) {
	cases := []struct {
		label string
		mock  func(usecase *usecases_mock.MockDriverUsecase)
		setup func() *http.Request
		code  int
		out   io.Reader
	}{
		{
			label: "success",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					SetState("foo", "bar").
					Return(nil).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString("OK\n"),
		},

		{
			label: "missing URL parameter",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing URL parameter \"name\"\n"),
		},

		{
			label: "missing X-Driver-Token header",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing X-Driver-Token header\n"),
		},

		{
			label: "token not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to set state for driver \"foo\": not found\n"),
		},

		{
			label: "unauthorized",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrUnauthorized).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusUnauthorized,
			out:  bytes.NewBufferString("failed to set state for driver \"foo\": unauthorized\n"),
		},

		{
			label: "internal authorization error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},

		{
			label: "missing Content-Type header",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("Bad Request\n"),
		},

		{
			label: "not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					SetState("foo", "bar").
					Return(lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to set state for driver \"foo\": not found\n"),
		},

		{
			label: "internal error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					SetState("foo", "bar").
					Return(lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/state", lib.MustJsonMarshalToBuffer(t, "bar"))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},
	}

	for _, tt := range cases {
		lib.RunCase(t, tt.label, func(t *testing.T) {
			failed := false

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := usecases_mock.NewMockDriverUsecase(ctrl)
			inject := func(context.Context) usecases.DriverUsecase { return usecase }
			controller := controllers.NewDriverController(inject)

			tt.mock(usecase)

			w := httptest.NewRecorder()
			r := tt.setup()

			b := &strings.Builder{}
			logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
			logger := log.Output(logout).Level(zerolog.TraceLevel)

			ctx := r.Context()
			ctx = logger.WithContext(ctx)

			controller.SetState(w, r.WithContext(ctx))

			if w.Code != tt.code {
				t.Errorf("%s %s got %d: expected %d", r.Method, r.RequestURI, w.Code, tt.code)
				failed = true
			}

			if ops := utils.ReaderDiff(w.Body, tt.out); ops != nil {
				t.Errorf("%s %s response body:\n%s", r.Method, r.RequestURI, utils.JoinOps(ops, "\n"))
				failed = true
			}

			if failed {
				t.Errorf("log output:\n%s", b.String())
			}
		})
	}
}

func TestDriverGetStatus(t *testing.T) {
	cases := []struct {
		label string
		mock  func(usecase *usecases_mock.MockDriverUsecase)
		setup func() *http.Request
		code  int
		out   io.Reader
	}{
		{
			label: "success",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					GetStatus("foo").
					Return(driver.Idle, nil).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/status", nil)
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusOK,
			out:  lib.MustJsonMarshalToBuffer(t, driver.Idle),
		},

		{
			label: "missing URL parameter",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/status", nil)
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing URL parameter \"name\"\n"),
		},

		{
			label: "not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					GetStatus("foo").
					Return(driver.Error, lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/status", nil)
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to get status for driver \"foo\": not found\n"),
		},

		{
			label: "internal error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					GetStatus("foo").
					Return(driver.Error, lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/status", nil)
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},
	}

	for _, tt := range cases {
		lib.RunCase(t, tt.label, func(t *testing.T) {
			failed := false

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := usecases_mock.NewMockDriverUsecase(ctrl)
			inject := func(context.Context) usecases.DriverUsecase { return usecase }
			controller := controllers.NewDriverController(inject)

			tt.mock(usecase)

			w := httptest.NewRecorder()
			r := tt.setup()

			b := &strings.Builder{}
			logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
			logger := log.Output(logout).Level(zerolog.TraceLevel)

			ctx := r.Context()
			ctx = logger.WithContext(ctx)

			controller.GetStatus(w, r.WithContext(ctx))

			if w.Code != tt.code {
				t.Errorf("%s %s got %d: expected %d", r.Method, r.RequestURI, w.Code, tt.code)
				failed = true
			}

			if ops := utils.ReaderDiff(w.Body, tt.out); ops != nil {
				t.Errorf("%s %s response body:\n%s", r.Method, r.RequestURI, utils.JoinOps(ops, "\n"))
				failed = true
			}

			if failed {
				t.Errorf("log output:\n%s", b.String())
			}
		})
	}
}

func TestDriverSetStatus(t *testing.T) {
	cases := []struct {
		label string
		mock  func(usecase *usecases_mock.MockDriverUsecase)
		setup func() *http.Request
		code  int
		out   io.Reader
	}{
		{
			label: "success",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					SetStatus("foo", driver.Idle).
					Return(nil).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString("OK\n"),
		},

		{
			label: "missing URL parameter",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing URL parameter \"name\"\n"),
		},

		{
			label: "missing X-Driver-Token header",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing X-Driver-Token header\n"),
		},

		{
			label: "token not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to set status for driver \"foo\": not found\n"),
		},

		{
			label: "unauthorized",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrUnauthorized).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusUnauthorized,
			out:  bytes.NewBufferString("failed to set status for driver \"foo\": unauthorized\n"),
		},

		{
			label: "internal authorization error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},

		{
			label: "missing Content-Type header",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("Bad Request\n"),
		},

		{
			label: "not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					SetStatus("foo", driver.Idle).
					Return(lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to set status for driver \"foo\": not found\n"),
		},

		{
			label: "internal error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					SetStatus("foo", driver.Idle).
					Return(lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPut, "/driver/foo/status", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},
	}

	for _, tt := range cases {
		lib.RunCase(t, tt.label, func(t *testing.T) {
			failed := false

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := usecases_mock.NewMockDriverUsecase(ctrl)
			inject := func(context.Context) usecases.DriverUsecase { return usecase }
			controller := controllers.NewDriverController(inject)

			tt.mock(usecase)

			w := httptest.NewRecorder()
			r := tt.setup()

			b := &strings.Builder{}
			logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
			logger := log.Output(logout).Level(zerolog.TraceLevel)

			ctx := r.Context()
			ctx = logger.WithContext(ctx)

			controller.SetStatus(w, r.WithContext(ctx))

			if w.Code != tt.code {
				t.Errorf("%s %s got %d: expected %d", r.Method, r.RequestURI, w.Code, tt.code)
				failed = true
			}

			if ops := utils.ReaderDiff(w.Body, tt.out); ops != nil {
				t.Errorf("%s %s response body:\n%s", r.Method, r.RequestURI, utils.JoinOps(ops, "\n"))
				failed = true
			}

			if failed {
				t.Errorf("log output:\n%s", b.String())
			}
		})
	}
}

func TestDriverOperation(t *testing.T) {
	cases := []struct {
		label string
		mock  func(usecase *usecases_mock.MockDriverUsecase)
		setup func() *http.Request
		code  int
		out   io.Reader
	}{
		{
			label: "success",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					GetOp("foo").
					Return(&driver.Op{
						Name: "op",
						Arg:  "arg",
					}, nil).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/operation", nil)
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusOK,
			out: lib.MustJsonMarshalToBuffer(t, driver.Op{
				Name: "op",
				Arg:  "arg",
			}),
		},

		{
			label: "missing URL parameter",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/operation", nil)
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing URL parameter \"name\"\n"),
		},

		{
			label: "missing X-Driver-Token header",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing X-Driver-Token header\n"),
		},

		{
			label: "token not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to get operation for driver \"foo\": not found\n"),
		},

		{
			label: "unauthorized",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrUnauthorized).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusUnauthorized,
			out:  bytes.NewBufferString("failed to get operation for driver \"foo\": unauthorized\n"),
		},

		{
			label: "internal authorization error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Idle))
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},

		{
			label: "not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					GetOp("foo").
					Return(&driver.Op{}, lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/operation", nil)
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to get operation for driver \"foo\": not found\n"),
		},

		{
			label: "internal error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					Authorize("foo", "foo").
					Return(nil).
					Times(1)
				usecase.EXPECT().
					GetOp("foo").
					Return(&driver.Op{}, lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodGet, "/driver/foo/operation", nil)
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},
	}

	for _, tt := range cases {
		lib.RunCase(t, tt.label, func(t *testing.T) {
			failed := false

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := usecases_mock.NewMockDriverUsecase(ctrl)
			inject := func(context.Context) usecases.DriverUsecase { return usecase }
			controller := controllers.NewDriverController(inject)

			tt.mock(usecase)

			w := httptest.NewRecorder()
			r := tt.setup()

			b := &strings.Builder{}
			logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
			logger := log.Output(logout).Level(zerolog.TraceLevel)

			ctx := r.Context()
			ctx = logger.WithContext(ctx)

			controller.Operation(w, r.WithContext(ctx))

			if w.Code != tt.code {
				t.Errorf("%s %s got %d: expected %d", r.Method, r.RequestURI, w.Code, tt.code)
				failed = true
			}

			if ops := utils.ReaderDiff(w.Body, tt.out); ops != nil {
				t.Errorf("%s %s response body:\n%s", r.Method, r.RequestURI, utils.JoinOps(ops, "\n"))
				failed = true
			}

			if failed {
				t.Errorf("log output:\n%s", b.String())
			}
		})
	}
}

func TestDriverDispatch(t *testing.T) {
	cases := []struct {
		label string
		mock  func(usecase *usecases_mock.MockDriverUsecase)
		setup func() *http.Request
		code  int
		out   io.Reader
	}{
		{
			label: "success",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					SetOp("foo", driver.Op{
						Name: "op",
						Arg:  "arg",
					}).
					Return(nil).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPost, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Op{
					Name: "op",
					Arg:  "arg",
				}))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusOK,
			out:  bytes.NewBufferString("OK\n"),
		},

		{
			label: "missing URL parameter",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				r := httptest.NewRequest(http.MethodPost, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Op{
					Name: "op",
					Arg:  "arg",
				}))
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("missing URL parameter \"name\"\n"),
		},

		{
			label: "missing Content-Type header",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPost, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Op{
					Name: "op",
					Arg:  "arg",
				}))
				r.Header.Set("X-Driver-Token", "foo")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("Bad Request\n"),
		},

		{
			label: "validation error",
			mock:  func(usecase *usecases_mock.MockDriverUsecase) {},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPost, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Op{
					Name: "",
					Arg:  nil,
				}))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusBadRequest,
			out:  bytes.NewBufferString("validation failed on field \"name\" for constraint \"required\"\n"),
		},

		{
			label: "not found",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					SetOp("foo", driver.Op{
						Name: "op",
						Arg:  "arg",
					}).
					Return(lib.ErrNotFound).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPost, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Op{
					Name: "op",
					Arg:  "arg",
				}))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusNotFound,
			out:  bytes.NewBufferString("failed to dispatch for driver \"foo\": not found\n"),
		},

		{
			label: "internal error",
			mock: func(usecase *usecases_mock.MockDriverUsecase) {
				usecase.EXPECT().
					SetOp("foo", driver.Op{
						Name: "op",
						Arg:  "arg",
					}).
					Return(lib.ErrUnknown).
					Times(1)
			},
			setup: func() *http.Request {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("name", "foo")
				r := httptest.NewRequest(http.MethodPost, "/driver/foo/operation", lib.MustJsonMarshalToBuffer(t, driver.Op{
					Name: "op",
					Arg:  "arg",
				}))
				r.Header.Set("X-Driver-Token", "foo")
				r.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
				return r.WithContext(ctx)
			},
			code: http.StatusInternalServerError,
			out:  bytes.NewBufferString("Internal Server Error\n"),
		},
	}

	for _, tt := range cases {
		lib.RunCase(t, tt.label, func(t *testing.T) {
			failed := false

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := usecases_mock.NewMockDriverUsecase(ctrl)
			inject := func(context.Context) usecases.DriverUsecase { return usecase }
			controller := controllers.NewDriverController(inject)

			tt.mock(usecase)

			w := httptest.NewRecorder()
			r := tt.setup()

			b := &strings.Builder{}
			logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
			logger := log.Output(logout).Level(zerolog.TraceLevel)

			ctx := r.Context()
			ctx = logger.WithContext(ctx)

			controller.Dispatch(w, r.WithContext(ctx))

			if w.Code != tt.code {
				t.Errorf("%s %s got %d: expected %d", r.Method, r.RequestURI, w.Code, tt.code)
				failed = true
			}

			if ops := utils.ReaderDiff(w.Body, tt.out); ops != nil {
				t.Errorf("%s %s response body:\n%s", r.Method, r.RequestURI, utils.JoinOps(ops, "\n"))
				failed = true
			}

			if failed {
				t.Errorf("log output:\n%s", b.String())
			}
		})
	}
}
