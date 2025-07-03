package logger

import (
	"fmt"

	"github.com/ratheeshkumar25/adsmetrictracker/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	Logger *zap.SugaredLogger
}

// NewProductionZapLogger will return a new production logger backed by zap
func NewLogger(cfg *config.Config) (*Logger, error) {
	conf := zap.NewProductionConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	conf.DisableCaller = true
	conf.DisableStacktrace = true
	fmt.Println(cfg.LogFile)

	//zapLogger, err := conf.Build(zap.WrapCore(zapCore))
	zapLogger, err := conf.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    1, // megabytes
			MaxBackups: 30,
			MaxAge:     30, // days
		})
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			w,
			zap.DebugLevel,
		)
		cores := zapcore.NewTee(c, core)
		return cores

	}))
	return &Logger{
		Logger: zapLogger.Sugar(),
	}, err
}
