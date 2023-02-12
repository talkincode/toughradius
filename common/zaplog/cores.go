package zaplog

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func (l *Logger) initConsoleCore() zapcore.Core {
	writer := zapcore.AddSync(os.Stdout)
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("[%v]", t.Format(time.RFC3339)))
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	logLevel := zapcore.DebugLevel
	if l.cfg.Mode == Prod {
		logLevel = zapcore.ErrorLevel
	}
	core := zapcore.NewCore(encoder, writer, logLevel)
	return core
}

func (l *Logger) initFileCore() zapcore.Core {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   l.cfg.Filename,
		MaxSize:    64,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   false,
	}
	writer := zapcore.AddSync(lumberJackLogger)
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("[%v]", t.Format(time.RFC3339)))
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	logLevel := zapcore.DebugLevel
	if l.cfg.Mode == Prod {
		logLevel = zapcore.ErrorLevel
	}
	core := zapcore.NewCore(encoder, writer, logLevel)
	return core
}

func (l *Logger) getLokiWriter() *lokiWriter {
	lw := &lokiWriter{}
	lw.client = NewLokiClient(l.cfg.LokiJob, l.cfg.LokiApi, l.cfg.LokiUser, l.cfg.LokiPwd, l.cfg.QueueSize)
	if os.Getenv("LOKI_CLIENT_DEBUG") == "true" {
		lw.client.Debug = true
	}
	go lw.client.Start()
	return lw
}

func (l *Logger) initLokiCore() zapcore.Core {
	l.lokiWriter = l.getLokiWriter()
	writer := zapcore.AddSync(l.lokiWriter)
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339))
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder // zapcore.EpochNanosTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, zapcore.DebugLevel)
}

func (l *Logger) initMetricsCore() zapcore.Core {
	l.metricsWriter = newMetricsWriter(l.cfg.QueueSize, l.cfg.MetricsStorage, time.Hour*time.Duration(l.cfg.MetricsHistory))
	writer := zapcore.AddSync(l.metricsWriter)
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339))
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder // zapcore.EpochNanosTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, zapcore.DebugLevel)
}
