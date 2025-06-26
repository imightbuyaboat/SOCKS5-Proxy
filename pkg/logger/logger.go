package logger

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once   sync.Once
	logger *zap.Logger
)

func InitLogger() {
	once.Do(func() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			pae.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z"))
		}

		var err error
		logger, err = config.Build()
		if err != nil {
			panic("cannot initialize zap logger: " + err.Error())
		}
	})
}

func GetLogger() *zap.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}
