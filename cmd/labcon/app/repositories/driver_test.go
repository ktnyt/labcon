package repositories_test

import (
	"errors"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/ktnyt/labcon/cmd/labcon/app/models"
	"github.com/ktnyt/labcon/cmd/labcon/app/repositories"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
	"github.com/ktnyt/labcon/utils"
)

func TestDriverCreate(t *testing.T) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()
	repo := repositories.NewDriverRepository(db)

	cases := []struct {
		name  string
		state interface{}
		token string
		err   error
	}{
		{
			name:  "foo",
			state: "foo",
			token: lib.Base32String(lib.NewToken(20)),
			err:   nil,
		},
		{
			name:  "bar",
			state: "bar",
			token: lib.Base32String(lib.NewToken(20)),
			err:   nil,
		},
		{
			name:  "foo",
			state: "bar",
			token: lib.Base32String(lib.NewToken(20)),
			err:   lib.ErrAlreadyExists,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			err := repo.Create(tt.name, tt.token, tt.state)
			if !errors.Is(err, tt.err) {
				t.Errorf("%T.Create(%q, token, state) = %v: expected %v", repo, tt.name, err, tt.err)
			}
		})
	}
}
func TestDriverFetch(t *testing.T) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()
	repo := repositories.NewDriverRepository(db)

	token := lib.Base32String(lib.NewToken(20))
	if err := repo.Create("foo", token, "foo"); err != nil {
		t.Fatalf("failed to create driver in fixture: %v", err)
	}

	cases := []struct {
		out models.DriverModel
		err error
	}{
		{
			out: models.NewDriver("foo", token, "foo"),
			err: nil,
		},
		{
			out: models.NewDriver("bar", token, "bar"),
			err: lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			out, err := repo.Fetch(tt.out.Name)
			if !errors.Is(err, tt.err) {
				t.Errorf("%T.Fetch(%q) = _, %v: expected %v", repo, tt.out.Name, err, tt.err)
			}

			if tt.err == nil {
				if ops := utils.ObjDiff(out, tt.out); ops != nil {
					t.Errorf("diff:\n%s", utils.JoinOps(ops, "\n"))
				}
			}
		})
	}
}

func TestDriverUpdate(t *testing.T) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()
	repo := repositories.NewDriverRepository(db)

	token := lib.Base32String(lib.NewToken(20))
	if err := repo.Create("foo", token, "foo"); err != nil {
		t.Fatalf("failed to create driver in fixture")
	}

	cases := []struct {
		name  string
		state interface{}
		err   error
	}{
		{
			name:  "foo",
			state: "bar",
			err:   nil,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			model, err := repo.Fetch(tt.name)
			if err != nil {
				t.Errorf("failed to fetch driver %q: %v", tt.name, err)
			}
			model.State = tt.state
			if err := repo.Update(model); !errors.Is(err, tt.err) {
				t.Errorf("%T.Update(%q) = %v: expected %v", repo, tt.name, err, tt.err)
			}

			if tt.err == nil {
				out, err := repo.Fetch(tt.name)
				if err != nil {
					t.Fatal("failed to fetch driver")
				}

				if ops := utils.ObjDiff(out.State, tt.state); ops != nil {
					t.Errorf("\n%s", utils.JoinOps(ops, "\n"))
				}
			}
		})
	}
}

func TestDriverDelete(t *testing.T) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()
	repo := repositories.NewDriverRepository(db)

	token := lib.Base32String(lib.NewToken(20))
	if err := repo.Create("foo", token, "foo"); err != nil {
		t.Fatalf("failed to create driver in fixture")
	}

	cases := []struct {
		name  string
		state interface{}
		err   error
	}{
		{
			name: "foo",
			err:  nil,
		},
		{
			name: "foo",
			err:  lib.ErrNotFound,
		},
	}

	for i, tt := range cases {
		lib.RunCase(t, i, func(t *testing.T) {
			err := repo.Delete(tt.name)
			if !errors.Is(err, tt.err) {
				t.Errorf("%T.Delete(%q) = %v: expected %v", repo, tt.name, err, tt.err)
			}

			if tt.err == nil {
				_, err := repo.Fetch(tt.name)
				if !errors.Is(err, lib.ErrNotFound) {
					t.Errorf("%T.Fetch(%q) = _, %v: expected %v", repo, tt.name, err, lib.ErrNotFound)
				}
			}
		})
	}
}
