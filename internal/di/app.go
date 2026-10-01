package di

import (
	appLayer "tictactoe/internal/app"
	"tictactoe/internal/handlers"
	"tictactoe/internal/handlers/middleware"
	jwtservice "tictactoe/internal/jwt"
	"tictactoe/internal/store"

	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		fx.Provide(
			store.NewStorage,
			store.NewRepository,
			appLayer.New,
			jwtservice.New,
			handlers.New,
			middleware.New,
			handlers.NewRouter,
		),
		fx.Invoke(startHTTPServer),
	)
}
