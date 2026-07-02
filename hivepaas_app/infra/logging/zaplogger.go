package logging

import (
	"go.uber.org/zap"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

type ZapLogger struct {
	Sync  func() error
	sugar *zap.SugaredLogger
}

func NewZapLogger(cfg *config.Config) (Logger, error) {
	var zapConfig zap.Config
	if !cfg.IsDevEnv() {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &ZapLogger{
		Sync:  logger.Sync,
		sugar: logger.Sugar(),
	}, nil
}

func (z *ZapLogger) Info(msg string, keysAndValues ...any) {
	z.sugar.Infow(msg, keysAndValues...)
}

func (z *ZapLogger) Error(msg string, keysAndValues ...any) {
	z.sugar.Errorw(msg, keysAndValues...)
}

func (z *ZapLogger) Debug(msg string, keysAndValues ...any) {
	z.sugar.Debugw(msg, keysAndValues...)
}

func (z *ZapLogger) Warn(msg string, keysAndValues ...any) {
	z.sugar.Warnw(msg, keysAndValues...)
}

func (z *ZapLogger) Infof(template string, args ...any) {
	z.sugar.Infof(template, args...)
}

func (z *ZapLogger) Errorf(template string, args ...any) {
	z.sugar.Errorf(template, args...)
}

func (z *ZapLogger) Warnf(template string, args ...any) {
	z.sugar.Warnf(template, args...)
}

func (z *ZapLogger) Debugf(template string, args ...any) {
	z.sugar.Debugf(template, args...)
}

func (z *ZapLogger) Fatal(keysAndValues ...any) {
	z.sugar.Fatal(keysAndValues)
}

func (z *ZapLogger) Panic(keysAndValues ...any) {
	z.sugar.Panic(keysAndValues...)
}

func (z *ZapLogger) Fatalf(template string, args ...any) {
	z.sugar.Fatalf(template, args...)
}

func (z *ZapLogger) Panicf(template string, args ...any) {
	z.sugar.Panicf(template, args...)
}
