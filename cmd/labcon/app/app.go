package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/ktnyt/labcon/cmd/labcon/app/controllers"
	"github.com/ktnyt/labcon/cmd/labcon/app/injectors"
	"github.com/ktnyt/labcon/cmd/labcon/app/views"
)

type App struct {
	driver controllers.DriverController
}

func NewApp(injectDriver injectors.DriverInjector) App {
	return App{
		driver: controllers.NewDriverController(injectDriver),
	}
}

func (a App) Setup(r chi.Router) {
	r.Get("/", views.EmptyView)
	r.Route("/driver", func(r chi.Router) {
		r.Post("/", a.driver.Register)
		r.Route("/{name}", func(r chi.Router) {
			r.Route("/state", func(r chi.Router) {
				r.Get("/", a.driver.GetState)
				r.Put("/", a.driver.SetState)
			})
			r.Route("/status", func(r chi.Router) {
				r.Get("/", a.driver.GetStatus)
				r.Put("/", a.driver.SetStatus)
			})
			r.Route("/operation", func(r chi.Router) {
				r.Get("/", a.driver.Operation)
				r.Post("/", a.driver.Dispatch)
			})
		})
	})
}
