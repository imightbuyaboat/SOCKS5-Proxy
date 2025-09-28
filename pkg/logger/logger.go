package logger

import (
	"bytes"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once      sync.Once
	logger    *zap.Logger
	logBuffer *bytes.Buffer
)

func InitLogger() {
	once.Do(func() {
		logBuffer = new(bytes.Buffer)

		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			pae.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z"))
		}

		consoleCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)

		bufferCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			zapcore.AddSync(logBuffer),
			zapcore.InfoLevel,
		)

		core := zapcore.NewTee(consoleCore, bufferCore)

		logger = zap.New(core, zap.AddCaller())
	})
}

func GetLogger() *zap.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}

func GetLogBuffer() string {
	if logBuffer == nil {
		return ""
	}
	return logBuffer.String()
}

func ClearLogBuffer() {
	if logBuffer != nil {
		logBuffer.Reset()
	}
}
