package log

import (
	"io"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Module = fx.Module("log",
	fx.Provide(
		fx.Annotate(
			NewLogWriter,
			fx.As(new(io.Writer)),
		),
	),
	fx.Provide(NewLogger),
)

func NewLogger(writer io.Writer) *zap.Logger {
	// default to stdout
	if writer == nil {
		writer = os.Stdout
	}

	encodeTime := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339))
	}

	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = encodeTime
	config.EncoderConfig.TimeKey = "@timestamp"

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		zapcore.AddSync(writer),
		config.Level.Level(),
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0)).WithLazy()
}

func NewLogWriter() io.Writer {
	return os.Stdout
}
