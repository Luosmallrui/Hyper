package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

func init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)
	L = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}
