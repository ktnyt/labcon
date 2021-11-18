package injectors

import (
	"context"

	"github.com/ktnyt/labcon/cmd/labcon/app/repositories"
	"github.com/ktnyt/labcon/cmd/labcon/app/usecases"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
)

type DriverInjector func(ctx context.Context) usecases.DriverUsecase

func Driver(ctx context.Context) usecases.DriverUsecase {
	generate := lib.UseDriverTokenGenerator(ctx)
	repository := repositories.NewDriverRepository(lib.UseBadger(ctx))
	usecase := usecases.NewDriverUsecase(repository, generate)
	return usecase
}
