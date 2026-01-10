package log

import (
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

func init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		projectName := "Hyper"

		index := strings.Index(caller.File, projectName)
		if index != -1 {
			enc.AppendString(caller.File[index:] + ":" + strconv.Itoa(caller.Line))
		} else {
			enc.AppendString(caller.TrimmedPath())
		}
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)
	L = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}
