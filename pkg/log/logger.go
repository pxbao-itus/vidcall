package log

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("log",
	fx.Provide(NewLogger),
)

func NewLogger() *zap.Logger {
	return zap.NewExample()
}
