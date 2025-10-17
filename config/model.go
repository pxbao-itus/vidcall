package config

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Config struct {
	HttpServer HttpServer
}

type ConfigParams struct {
	fx.In

	Logger *zap.Logger
}

type ConfigResult struct {
	fx.Out

	Config Config
}

type HttpServer struct {
	Host string
	Port string
}

func (h HttpServer) ToAddr() string {
	return h.Host + ":" + h.Port
}
