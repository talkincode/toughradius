package zaplog

import (
	"github.com/nakabonne/tstorage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Dev  string = "development"
	Prod string = "production"
)

type Logger struct {
	cfg           LogConfig
	lokiWriter    *lokiWriter
	metricsWriter *metricsWriter
	initialized   bool
}

var logger *Logger

func init() {
	logger = &Logger{}
	logger.Init(LogConfig{
		Mode:       Dev,
		LokiEnable: false,
		FileEnable: false,
	})
}

func (l *Logger) Init(c LogConfig) {
	l.cfg = c
	var cores []zapcore.Core
	consoleCore := l.initConsoleCore()
	cores = append(cores, consoleCore)
	// 生成输出到Loki的Core
	if c.LokiEnable {
		lokiCore := l.initLokiCore()
		cores = append(cores, lokiCore)
	}
	// 输出到文件的Core
	if c.FileEnable {
		fileCore := l.initFileCore()
		cores = append(cores, fileCore)
	}

	metricsCore := l.initMetricsCore()
	cores = append(cores, metricsCore)

	_logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
	l.initialized = true
	zap.ReplaceGlobals(_logger)
}

func InitGlobalLogger(c LogConfig) {
	logger.Init(c)
}

func TSDB() tstorage.Storage {
	return logger.metricsWriter.tsdb
}

//
// func GetLogger(c LogConfig) *zap.Logger {
// 	l := &Logger{}
// 	l.cfg = c
// 	var cores []zapcore.Core
// 	if l.cfg.ConsoleEnable {
// 		consoleCore := l.initConsoleCore()
// 		cores = append(cores, consoleCore)
// 	}
// 	// 生成输出到Loki的Core
// 	if c.LokiEnable {
// 		lokiCore := l.initLokiCore()
// 		cores = append(cores, lokiCore)
// 	}
// 	// 输出到文件的Core
// 	if c.FileEnable {
// 		fileCore := l.initFileCore()
// 		cores = append(cores, fileCore)
// 	}
//
// 	metricsCore := l.initMetricsCore()
// 	cores = append(cores, metricsCore)
//
// 	_logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
// 	l.initialized = true
// 	return _logger
// }

func (l *Logger) Release() {
	logger.lokiWriter.client.Stop()
	if logger.metricsWriter.tsdb != nil {
		_ = logger.metricsWriter.tsdb.Close()
	}
}

func Release() {
	logger.Release()
}
