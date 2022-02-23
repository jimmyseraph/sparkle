package logger

import (
	"fmt"

	"github.com/jimmyseraph/sparkle/engine"
	"github.com/jimmyseraph/sparkle/utils/logger"

	"go.uber.org/zap"
)

var _ engine.Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	log *zap.SugaredLogger
}

func NewZapLogger() *ZapLogger {
	log := logger.NewLogger("", "info")
	return &ZapLogger{log: log}
}

func (z *ZapLogger) Log(logType string, message string, args ...interface{}) {

	z.log.Info(logType + ": " + fmt.Sprintf(message, args...))
}
