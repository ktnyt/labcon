package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ktnyt/labcon/cmd/labcon/app/usecases"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
	"github.com/ktnyt/labcon/driver"
)

type DriverController interface {
	List(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	GetState(w http.ResponseWriter, r *http.Request)
	SetState(w http.ResponseWriter, r *http.Request)
	GetStatus(w http.ResponseWriter, r *http.Request)
	SetStatus(w http.ResponseWriter, r *http.Request)
	Operation(w http.ResponseWriter, r *http.Request)
	Dispatch(w http.ResponseWriter, r *http.Request)
	Disconnect(w http.ResponseWriter, r *http.Request)
}

type DriverControllerImpl struct {
	inject func(context.Context) usecases.DriverUsecase
}

func NewDriverController(inject func(context.Context) usecases.DriverUsecase) DriverController {
	return DriverControllerImpl{inject: inject}
}

func (controller DriverControllerImpl) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	list, err := usecase.List()
	if err != nil {
		logger.Err(err).Msgf("failed to list drivers")
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.JsonResponse(w, ctx, list)
}

func (controller DriverControllerImpl) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	var req driver.RegisterParams
	if err := lib.JsonRequest(r, &req); err != nil {
		logger.Warn().Err(err).Msg("failed to process request")
		lib.HTTPError(w, http.StatusBadRequest)
		return
	}

	if err := lib.Validate(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := usecase.Register(req.Name, req.State)
	if err != nil {
		if errors.Is(err, lib.ErrAlreadyExists) {
			http.Error(w, fmt.Sprintf("failed to register driver %q: %v", req.Name, err), http.StatusBadRequest)
			return
		}
		logger.Err(err).Msgf("failed to register driver %q", req.Name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.JsonResponse(w, ctx, token)
}

func (controller DriverControllerImpl) GetState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "missing URL parameter \"name\"", http.StatusBadRequest)
		return
	}

	state, err := usecase.GetState(name)
	if err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to get state for driver %q: %v", name, err), http.StatusNotFound)
			return
		}
		logger.Err(err).Msgf("failed to get state for driver %q", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.JsonResponse(w, ctx, state)
}

func (controller DriverControllerImpl) SetState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "missing URL parameter \"name\"", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("X-Driver-Token")
	if token == "" {
		http.Error(w, "missing X-Driver-Token header", http.StatusUnauthorized)
		return
	}

	if err := usecase.Authorize(name, token); err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to authorize driver %q in set state: %v", name, err), http.StatusNotFound)
			return
		}
		if errors.Is(err, lib.ErrForbidden) {
			http.Error(w, fmt.Sprintf("failed to authorize driver %q in set state: %v", name, err), http.StatusForbidden)
			return
		}
		logger.Err(err).Msgf("failed to authorize driver %q in set state", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	var state interface{}
	if err := lib.JsonRequest(r, &state); err != nil {
		logger.Warn().Err(err).Msg("failed to process request")
		lib.HTTPError(w, http.StatusBadRequest)
		return
	}

	if err := usecase.SetState(name, state); err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to set state for driver %q: %v", name, err), http.StatusNotFound)
			return
		}
		logger.Err(err).Msgf("failed to set state for driver %q", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.HTTPError(w, http.StatusOK)
}

func (controller DriverControllerImpl) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "missing URL parameter \"name\"", http.StatusBadRequest)
		return
	}

	status, err := usecase.GetStatus(name)
	if err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to get status for driver %q: %v", name, err), http.StatusNotFound)
			return
		}
		logger.Err(err).Msgf("failed to get status for driver %q", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.JsonResponse(w, ctx, status)
}

func (controller DriverControllerImpl) SetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "missing URL parameter \"name\"", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("X-Driver-Token")
	if token == "" {
		http.Error(w, "missing X-Driver-Token header", http.StatusUnauthorized)
		return
	}

	if err := usecase.Authorize(name, token); err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to authorize driver %q in set status: %v", name, err), http.StatusNotFound)
			return
		}
		if errors.Is(err, lib.ErrForbidden) {
			http.Error(w, fmt.Sprintf("failed to authorize driver %q in set status: %v", name, err), http.StatusForbidden)
			return
		}
		logger.Err(err).Msgf("failed to authorize driver %q in set status", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	var status driver.Status
	if err := lib.JsonRequest(r, &status); err != nil {
		logger.Warn().Err(err).Msg("failed to process request")
		lib.HTTPError(w, http.StatusBadRequest)
		return
	}

	if err := usecase.SetStatus(name, status); err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to set status for driver %q: %v", name, err), http.StatusNotFound)
			return
		}
		logger.Err(err).Msgf("failed to set status for driver %q", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.HTTPError(w, http.StatusOK)
}

func (controller DriverControllerImpl) Operation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "missing URL parameter \"name\"", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("X-Driver-Token")
	if token == "" {
		http.Error(w, "missing X-Driver-Token header", http.StatusUnauthorized)
		return
	}

	if err := usecase.Authorize(name, token); err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to authorize driver %q in get operation: %v", name, err), http.StatusNotFound)
			return
		}
		if errors.Is(err, lib.ErrForbidden) {
			http.Error(w, fmt.Sprintf("failed to authorize driver %q in get operation: %v", name, err), http.StatusForbidden)
			return
		}
		logger.Err(err).Msgf("failed to authorize driver %q in get operation", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	op, err := usecase.GetOp(name)
	if err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to get operation for driver %q: %v", name, err), http.StatusNotFound)
			return
		}
		logger.Err(err).Msgf("failed to get operation for driver %q", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.JsonResponse(w, ctx, op)
}

func (controller DriverControllerImpl) Dispatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "missing URL parameter \"name\"", http.StatusBadRequest)
		return
	}

	var op driver.Op
	if err := lib.JsonRequest(r, &op); err != nil {
		logger.Warn().Err(err).Msg("failed to process request")
		lib.HTTPError(w, http.StatusBadRequest)
		return
	}

	if err := lib.Validate(op); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := usecase.SetOp(name, op); err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to dispatch for driver %q: %v", name, err), http.StatusNotFound)
			return
		}
		logger.Err(err).Msgf("failed to dispatch for driver %q", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.HTTPError(w, http.StatusOK)
}

func (controller DriverControllerImpl) Disconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := lib.UseLogger(ctx)

	// Dependency injection.
	usecase := controller.inject(ctx)

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "missing URL parameter \"name\"", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("X-Driver-Token")
	if token == "" {
		http.Error(w, "missing X-Driver-Token header", http.StatusUnauthorized)
		return
	}

	if err := usecase.Authorize(name, token); err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to authorize driver %q in disconnect: %v", name, err), http.StatusNotFound)
			return
		}
		if errors.Is(err, lib.ErrForbidden) {
			http.Error(w, fmt.Sprintf("failed to authorize driver %q in disconnect: %v", name, err), http.StatusForbidden)
			return
		}
		logger.Err(err).Msgf("failed to authorize driver %q in disconnect", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	err := usecase.Delete(name)
	if err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed to disconnect driver %q: %v", name, err), http.StatusNotFound)
			return
		}
		logger.Err(err).Msgf("failed to disconnect driver %q", name)
		lib.HTTPError(w, http.StatusInternalServerError)
		return
	}

	lib.HTTPError(w, http.StatusOK)
}
