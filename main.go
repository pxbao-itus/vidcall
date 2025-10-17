package main

import (
	"vidcall/app"
	"vidcall/config"
	"vidcall/internal/module/room"
	"vidcall/internal/module/view"
	"vidcall/pkg/log"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),

		config.Module,
		log.Module,
		app.Module,
		room.Module,
		view.Module,

		fx.Invoke(app.Invoke()),
	).Run()
}
