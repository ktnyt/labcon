package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrJsonContentType = errors.New("Content-Type is not application/json")
	ErrJsonHTTPMethod  = errors.New("HTTP method does not have a body")
)

func JsonRequest(r *http.Request, p interface{}) error {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if r.Header.Get("Content-Type") != "application/json" {
			return ErrJsonContentType
		}
	default:
		return ErrJsonHTTPMethod
	}

	return json.NewDecoder(r.Body).Decode(p)
}

func JsonResponse(w http.ResponseWriter, ctx context.Context, p interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(p); err != nil {
		logger := UseLogger(ctx)
		logger.Error().Err(err).Msg("failed to process response")
		HTTPError(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}
