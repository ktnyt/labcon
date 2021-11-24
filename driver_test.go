package labcon

import (
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

func TestDriver(t *testing.T) {
	r := chi.NewMux()

	b := &strings.Builder{}
	logout := zerolog.ConsoleWriter{Out: b, TimeFormat: time.RFC3339}
	logger := log.Output(logout).Level(zerolog.TraceLevel)

	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	r.Use(
		lib.Logger(logger),
		lib.Badger(db),
		lib.DriverTokenGenerator(lib.DefaultTokenGenerator),
	)

	a := app.NewApp(injectors.Driver)
	a.Setup(r)

	server := httptest.NewServer(r)
	defer server.Close()

	client := NewClient(server.URL)
	d, err := NewDriver(client, "foo", "foo")
	if err != nil {
		t.Fatal(err)
	}

	var state string
	if err := d.GetState(&state); err != nil {
		t.Fatal(err)
	}

	if state != "foo" {
		t.Fatalf("client state = %q, want \"foo\"", state)
	}

	status, err := d.GetStatus()
	if err != nil {
		t.Fatal(err)
	}

	if status != driver.Idle {
		t.Fatalf("client status = %q, want %q", status, driver.Idle)
	}

	op, err := d.Operation()
	if err != nil {
		t.Fatal(err)
	}

	if op != nil {
		t.Fatalf("client op = %v, want nil", op)
	}

	if err := d.Dispatch(driver.Op{
		Name: "op",
		Arg:  "arg",
	}); err != nil {
		t.Fatal(err)
	}

	op, err = d.Operation()
	if err != nil {
		t.Fatal(err)
	}

	if ops := utils.ObjDiff(op, driver.Op{
		Name: "op",
		Arg:  "arg",
	}); ops != nil {
		t.Fatal(utils.JoinOps(ops, "\n"))
	}

	status, err = d.GetStatus()
	if err != nil {
		t.Fatal(err)
	}

	if status != driver.Busy {
		t.Fatalf("client status = %q, want %q", status, driver.Busy)
	}

	if err := d.SetState("bar"); err != nil {
		t.Fatal(err)
	}

	if err := d.GetState(&state); err != nil {
		t.Fatal(err)
	}

	if state != "bar" {
		t.Fatalf("client state = %q, want \"bar\"", state)
	}

	if err := d.SetStatus(driver.Idle); err != nil {
		t.Fatal(err)
	}

	status, err = d.GetStatus()
	if err != nil {
		t.Fatal(err)
	}

	if status != driver.Idle {
		t.Fatalf("client status = %q, want %q", status, driver.Idle)
	}

	op, err = d.Operation()
	if err != nil {
		t.Fatal(err)
	}

	if op != nil {
		t.Fatalf("client op = %v, want nil", op)
	}
}
