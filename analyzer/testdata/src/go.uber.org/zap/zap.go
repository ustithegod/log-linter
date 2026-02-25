package zap

import "go.uber.org/zap/zapcore"

type Logger struct{}

type Field struct{}
type Option struct{}

func NewProduction(opts ...Option) (*Logger, error) {
	return &Logger{}, nil
}

func (l *Logger) Info(msg string, fields ...Field)  {}
func (l *Logger) Warn(msg string, fields ...Field)  {}
func (l *Logger) Error(msg string, fields ...Field) {}
func (l *Logger) Log(level zapcore.Level, msg string, fields ...Field) {
}
