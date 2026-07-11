package logger

import "go.uber.org/zap"

type Logger struct {
	logger *zap.Logger
}

func NewLogger(kind string) (*Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return &Logger{
		logger: logger,
	}, nil
}

func NewNopLogger() *Logger {
	return &Logger{logger: zap.NewNop()}
}

func (l *Logger) Sync() error {
	return l.logger.Sync()
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}
