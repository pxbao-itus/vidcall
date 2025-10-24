package app

import (
	"context"
	"errors"
	"net/http"

	"vidcall/app/rest"
	"vidcall/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"server",
	rest.Module,
	fx.Provide(NewServer),
)

type Server struct {
	server *http.Server
	logger *zap.Logger
	config config.HttpServer
}

type ServerParams struct {
	fx.In

	Logger  *zap.Logger
	Config  config.Config
	Handler http.Handler
}

type ServerResult struct {
	fx.Out

	Server *Server
}

func newServer(params ServerParams) ServerResult {
	srv := &Server{
		server: &http.Server{
			Addr:    params.Config.HttpServer.ToAddr(),
			Handler: params.Handler,
		},
		logger: params.Logger,
		config: params.Config.HttpServer,
	}

	return ServerResult{
		Server: srv,
	}
}

func NewServer(lc fx.Lifecycle, params ServerParams) ServerResult {
	result := newServer(params)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go result.Server.Start(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			result.Server.Shutdown(ctx)
			return nil
		},
	})
	return result
}

func Invoke() func(server *Server) {
	return func(srv *Server) {}
}

func (s *Server) Start(ctx context.Context) {
	protocol := "HTTP"
	if s.config.UseHTTPS {
		protocol = "HTTPS"
	}

	s.logger.Info("Starting server",
		zap.String("protocol", protocol),
		zap.String("addr", s.server.Addr),
	)

	var err error
	if s.config.UseHTTPS {
		s.logger.Info("Using HTTPS",
			zap.String("cert", s.config.CertFile),
			zap.String("key", s.config.KeyFile),
		)
		err = s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
	} else {
		err = s.server.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatal("Server failed", zap.Error(err))
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Info("Shutting down server")
	if err := s.server.Close(); err != nil {
		s.logger.Error("Failed to shut down server", zap.Error(err))
	}
}
