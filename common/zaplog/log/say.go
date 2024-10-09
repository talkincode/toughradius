package log

import "go.uber.org/zap"

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Debug(args...)
}

func Debug2(msg string, fields ...zap.Field) {
	defer zap.L().Sync()
	zap.L().Debug(msg, fields...)
}

// Info uses fmt.Sprint to construct and log a message.
func Info(args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Info(args...)
}

func Info2(msg string, fields ...zap.Field) {
	defer zap.L().Sync()
	zap.L().Info(msg, fields...)
}

func Warn(args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Warn(args...)
}

func WarnDetail(msg string, fields ...zap.Field) {
	defer zap.L().Sync()
	zap.L().Warn(msg, fields...)
}

// Error uses fmt.Sprint to construct and log a message.
func Error(args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Error(args...)
}

func ErrorDetail(msg string, fields ...zap.Field) {
	defer zap.L().Sync()
	zap.L().Error(msg, fields...)
}

func ErrorIf(args interface{}) {
	if args == nil {
		return
	}
	defer zap.S().Sync()
	zap.S().Error(args)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Fatal(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Errorf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(template string, args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	defer zap.S().Sync()
	zap.S().Fatalf(template, args...)
}
