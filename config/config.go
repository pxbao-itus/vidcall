package config

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("config",
	fx.Provide(NewConfig),
)

func NewConfig(params ConfigParams) ConfigResult {
	params.Logger.Info("Loading configuration...")

	config := Config{
		HttpServer: HttpServer{
			Port: "8080",
		},
	}

	params.Logger.Info("Loaded configuration", zap.Any("config", config))

	return ConfigResult{
		Config: config,
	}
}
