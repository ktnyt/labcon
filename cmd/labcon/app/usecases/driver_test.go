package usecases_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ktnyt/labcon/cmd/labcon/app/models"
	"github.com/ktnyt/labcon/cmd/labcon/app/repositories_mock"
	"github.com/ktnyt/labcon/cmd/labcon/app/usecases"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
	"github.com/ktnyt/labcon/driver"
	"github.com/ktnyt/labcon/utils"
)

func TestDriverRegister(t *testing.T) {
	token := lib.Base32String(lib.NewToken(20))

	cases := []struct {
		mock func(repository *repositories_mock.MockDriverRepository)
		err  error
	}{
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Create("foo", token, "foo").
					Return(nil).
					Times(1)
			},
			err: nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Create("foo", token, "foo").
					Return(lib.ErrNotFound).
					Times(1)
			},
			err: lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repository := repositories_mock.NewMockDriverRepository(ctrl)
			tt.mock(repository)

			usecase := usecases.NewDriverUsecase(repository, func() string { return token })
			out, err := usecase.Register("foo", "foo")

			if out != token || !errors.Is(err, tt.err) {
				t.Errorf("usecase.Register(\"foo\", \"foo\") = (%s, %v): expected (%s, %v)", out, err, token, tt.err)
			}
		})
	}
}

func TestDriverAuthorize(t *testing.T) {
	cases := []struct {
		mock func(repository *repositories_mock.MockDriverRepository)
		err  error
	}{
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:  "foo",
						Token: "foo",
					}, nil).
					Times(1)
			},
			err: nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:  "foo",
						Token: "bar",
					}, nil).
					Times(1)
			},
			err: lib.ErrUnauthorized,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{}, lib.ErrNotFound).
					Times(1)
			},
			err: lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repository := repositories_mock.NewMockDriverRepository(ctrl)
			tt.mock(repository)

			usecase := usecases.NewDriverUsecase(repository, func() string { return "" })
			err := usecase.Authorize("foo", "foo")

			if !errors.Is(err, tt.err) {
				t.Errorf("%T.Authorize(\"foo\", \"foo\") = (_, %v): expected (_, %v)", usecase, err, tt.err)
			}
		})
	}
}

func TestDriverGetState(t *testing.T) {
	token := lib.Base32String(lib.NewToken(20))

	cases := []struct {
		mock func(repository *repositories_mock.MockDriverRepository)
		out  interface{}
		err  error
	}{
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:  "foo",
						State: "foo",
					}, nil).
					Times(1)
			},
			out: "foo",
			err: nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{}, lib.ErrNotFound).
					Times(1)
			},
			out: nil,
			err: lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repository := repositories_mock.NewMockDriverRepository(ctrl)
			tt.mock(repository)

			usecase := usecases.NewDriverUsecase(repository, func() string { return token })
			out, err := usecase.GetState("foo")

			if !errors.Is(err, tt.err) {
				t.Errorf("%T.GetState(\"foo\") = (_, %v): expected (_, %v)", usecase, err, tt.err)
			}

			if tt.err == nil {
				if ops := utils.ObjDiff(out, tt.out); ops != nil {
					t.Errorf("%T.GetState(\"foo\"):\n%s", usecase, ops)
				}
			}
		})
	}
}

func And(flags ...bool) bool {
	for _, flag := range flags {
		if !flag {
			return false
		}
	}
	return true
}

type driverModelMatcher models.DriverModel

func DriverModelMatcher(driver models.DriverModel) driverModelMatcher {
	return driverModelMatcher(driver)
}

func (matcher driverModelMatcher) Matches(arg interface{}) bool {
	driver, ok := arg.(models.DriverModel)
	return ok && And(
		driver.Name == matcher.Name,
		driver.Token == matcher.Token,
		driver.Status == matcher.Status,
		reflect.DeepEqual(driver.State, matcher.State),
		reflect.DeepEqual(driver.Op, matcher.Op),
	)
}

func (matcher driverModelMatcher) String() string {
	return fmt.Sprintf(
		"name = %q, token = %q, state = %v, status = %v, op = %v",
		matcher.Name, matcher.Token, matcher.State, matcher.Status, matcher.Op,
	)
}

func TestDriverSetState(t *testing.T) {
	token := lib.Base32String(lib.NewToken(20))

	cases := []struct {
		mock func(repository *repositories_mock.MockDriverRepository)
		err  error
	}{
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:  "foo",
						State: "foo",
					}, nil).
					Times(1)
				repository.EXPECT().
					Update(DriverModelMatcher(models.DriverModel{
						Name:  "foo",
						State: "bar",
					})).
					Return(nil).
					Times(1)
			},
			err: nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{}, lib.ErrNotFound).
					Times(1)
			},
			err: lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repository := repositories_mock.NewMockDriverRepository(ctrl)
			tt.mock(repository)

			usecase := usecases.NewDriverUsecase(repository, func() string { return token })
			err := usecase.SetState("foo", "bar")

			if !errors.Is(err, tt.err) {
				t.Errorf("%T.SetState(\"foo\", \"bar\") = (_, %v): expected (_, %v)", usecase, err, tt.err)
			}
		})
	}
}

func TestDriverGetStatus(t *testing.T) {
	token := lib.Base32String(lib.NewToken(20))

	cases := []struct {
		mock func(repository *repositories_mock.MockDriverRepository)
		out  driver.Status
		err  error
	}{
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:   "foo",
						Status: driver.Idle,
					}, nil).
					Times(1)
			},
			out: driver.Idle,
			err: nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{}, lib.ErrNotFound).
					Times(1)
			},
			out: driver.Error,
			err: lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repository := repositories_mock.NewMockDriverRepository(ctrl)
			tt.mock(repository)

			usecase := usecases.NewDriverUsecase(repository, func() string { return token })
			out, err := usecase.GetStatus("foo")

			if !errors.Is(err, tt.err) {
				t.Errorf("%T.GetStatus(\"foo\") = (_, %v): expected (_, %v)", usecase, err, tt.err)
			}

			if tt.err == nil {
				if out != tt.out {
					t.Errorf("%T.GetStatus(\"foo\") = (%v, nil): expected (%v, nil)", usecase, out, tt.out)
				}
			}
		})
	}
}

func TestDriverSetStatus(t *testing.T) {
	token := lib.Base32String(lib.NewToken(20))

	cases := []struct {
		mock   func(repository *repositories_mock.MockDriverRepository)
		status driver.Status
		err    error
	}{
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:   "foo",
						Status: driver.Idle,
					}, nil).
					Times(1)
				repository.EXPECT().
					Update(DriverModelMatcher(models.DriverModel{
						Name:   "foo",
						Status: driver.Busy,
					})).
					Return(nil).
					Times(1)
			},
			status: driver.Busy,
			err:    nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:   "foo",
						Status: driver.Busy,
						Op: &driver.Op{
							Name: "op",
							Arg:  "arg",
						},
					}, nil).
					Times(1)
				repository.EXPECT().
					Update(DriverModelMatcher(models.DriverModel{
						Name:   "foo",
						Status: driver.Idle,
					})).
					Return(nil).
					Times(1)
			},
			status: driver.Idle,
			err:    nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{}, lib.ErrNotFound).
					Times(1)
			},
			status: driver.Idle,
			err:    lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repository := repositories_mock.NewMockDriverRepository(ctrl)
			tt.mock(repository)

			usecase := usecases.NewDriverUsecase(repository, func() string { return token })
			err := usecase.SetStatus("foo", tt.status)

			if !errors.Is(err, tt.err) {
				t.Errorf("%T.SetStatus(\"foo\", \"%v\") = %v: expected %v", usecase, tt.status, err, tt.err)
			}
		})
	}
}

func TestDriverGetOp(t *testing.T) {
	token := lib.Base32String(lib.NewToken(20))

	cases := []struct {
		mock func(repository *repositories_mock.MockDriverRepository)
		op   *driver.Op
		err  error
	}{
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:  "foo",
						Token: token,
						State: "foo",
						Op: &driver.Op{
							Name: "op",
							Arg:  "arg",
						},
					}, nil).
					Times(1)
			},
			op: &driver.Op{
				Name: "op",
				Arg:  "arg",
			},
			err: nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{}, lib.ErrNotFound).
					Times(1)
			},
			op:  nil,
			err: lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repository := repositories_mock.NewMockDriverRepository(ctrl)
			tt.mock(repository)

			usecase := usecases.NewDriverUsecase(repository, func() string { return token })
			out, err := usecase.GetOp("foo")

			if !errors.Is(err, tt.err) {
				t.Errorf("%T.GetOp(\"foo\") = (_, %v): expected (_, %v)", usecase, err, tt.err)
			}

			if tt.err == nil {
				if ops := utils.ObjDiff(out, tt.op); ops != nil {
					t.Error(utils.JoinOps(ops, "\n"))
				}
			}
		})
	}
}

func TestDriverSetOp(t *testing.T) {
	token := lib.Base32String(lib.NewToken(20))

	cases := []struct {
		mock func(repository *repositories_mock.MockDriverRepository)
		err  error
	}{
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:   "foo",
						Token:  token,
						State:  "foo",
						Status: driver.Idle,
						Op:     nil,
					}, nil).
					Times(1)
				repository.EXPECT().
					Update(models.DriverModel{
						Name:   "foo",
						Token:  token,
						State:  "foo",
						Status: driver.Busy,
						Op: &driver.Op{
							Name: "op",
							Arg:  "arg",
						},
					}).
					Return(nil).
					Times(1)
			},
			err: nil,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{}, lib.ErrNotFound).
					Times(1)
			},
			err: lib.ErrNotFound,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:   "foo",
						Token:  token,
						State:  "foo",
						Status: driver.Busy,
						Op:     nil,
					}, nil).
					Times(1)
			},
			err: lib.ErrAlreadyExists,
		},
		{
			mock: func(repository *repositories_mock.MockDriverRepository) {
				repository.EXPECT().
					Fetch("foo").
					Return(models.DriverModel{
						Name:   "foo",
						Token:  token,
						State:  "foo",
						Status: driver.Idle,
						Op: &driver.Op{
							Name: "op",
							Arg:  "arg",
						},
					}, nil).
					Times(1)
			},
			err: lib.ErrAlreadyExists,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repository := repositories_mock.NewMockDriverRepository(ctrl)
			tt.mock(repository)

			usecase := usecases.NewDriverUsecase(repository, func() string { return token })
			err := usecase.SetOp("foo", driver.Op{
				Name: "op",
				Arg:  "arg",
			})

			if !errors.Is(err, tt.err) {
				t.Errorf("%T.SetOp(\"foo\", op) = %v: expected %v", usecase, err, tt.err)
			}
		})
	}
}
