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
			Host:     "0.0.0.0",
			Port:     "8080",
			UseHTTPS: true,
			CertFile: "certs/cert.pem",
			KeyFile:  "certs/key.pem",
		},
	}

	params.Logger.Info("Loaded configuration", zap.Any("config", config))

	return ConfigResult{
		Config: config,
	}
}
