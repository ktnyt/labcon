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

func TestClient(t *testing.T) {
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

	token, err := client.Register("foo", "foo")
	if err != nil {
		t.Fatal(err)
	}

	var state string
	if err := client.GetState("foo", &state); err != nil {
		t.Fatal(err)
	}

	if state != "foo" {
		t.Fatalf("client state = %q, want \"foo\"", state)
	}

	status, err := client.GetStatus("foo")
	if err != nil {
		t.Fatal(err)
	}

	if status != driver.Idle {
		t.Fatalf("client status = %q, want %q", status, driver.Idle)
	}

	op, err := client.Operation("foo", token)
	if err != nil {
		t.Fatal(err)
	}

	if op != nil {
		t.Fatalf("client op = %v, want nil", op)
	}

	if err := client.Dispatch("foo", driver.Op{
		Name: "op",
		Arg:  "arg",
	}); err != nil {
		t.Fatal(err)
	}

	op, err = client.Operation("foo", token)
	if err != nil {
		t.Fatal(err)
	}

	if ops := utils.ObjDiff(op, driver.Op{
		Name: "op",
		Arg:  "arg",
	}); ops != nil {
		t.Fatal(utils.JoinOps(ops, "\n"))
	}

	status, err = client.GetStatus("foo")
	if err != nil {
		t.Fatal(err)
	}

	if status != driver.Busy {
		t.Fatalf("client status = %q, want %q", status, driver.Busy)
	}

	if err := client.SetState("foo", token, "bar"); err != nil {
		t.Fatal(err)
	}

	if err := client.GetState("foo", &state); err != nil {
		t.Fatal(err)
	}

	if state != "bar" {
		t.Fatalf("client state = %q, want \"bar\"", state)
	}

	if err := client.SetStatus("foo", token, driver.Idle); err != nil {
		t.Fatal(err)
	}

	status, err = client.GetStatus("foo")
	if err != nil {
		t.Fatal(err)
	}

	if status != driver.Idle {
		t.Fatalf("client status = %q, want %q", status, driver.Idle)
	}

	op, err = client.Operation("foo", token)
	if err != nil {
		t.Fatal(err)
	}

	if op != nil {
		t.Fatalf("client op = %v, want nil", op)
	}
}
