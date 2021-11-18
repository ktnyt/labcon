package views

import (
	"net/http"

	"github.com/ktnyt/labcon/cmd/labcon/lib"
)

func EmptyView(w http.ResponseWriter, r *http.Request) {
	lib.HTTPError(w, http.StatusOK)
}
