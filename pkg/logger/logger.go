package logger

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dv-net/mx/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type sugaredLogger = zap.SugaredLogger

type Logger interface {
	logger.ExtendedLogger
	WithDBSyncer(dbSyncer *DBWriteSyncer)

	LastLogs() []MemoryLogDTO
}

type MemoryLogDTO struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

type WrappedLogger struct {
	logger   logger.Logger
	dbSyncer *DBWriteSyncer

	*sugaredLogger
	// log buffer
	mu       sync.Mutex
	logs     []MemoryLogDTO
	capacity int
	minLevel zapcore.Level
}
type LogPrams struct {
	Status    LogStatus
	Slug      string
	ProcessID uuid.UUID
}

var _ Logger = (*WrappedLogger)(nil)

func New(appVersion string, conf logger.Config) Logger {
	l := logger.NewExtended(
		logger.WithLogFormat(logger.LoggerFormatJSON),
		logger.WithAppVersion(appVersion),
		logger.WithConfig(conf),
	)

	return &WrappedLogger{
		logger:        l,
		sugaredLogger: l.Sugar(),
		capacity:      LogBufferSize,
		logs:          make([]MemoryLogDTO, 0, LogBufferSize),
		minLevel:      safeLevel(conf.Level),
	}
}

func (l *WrappedLogger) addToMemory(level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logs == nil {
		l.logs = make([]MemoryLogDTO, 0, l.capacity)
	}

	newLog := MemoryLogDTO{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	}

	l.logs = append([]MemoryLogDTO{newLog}, l.logs...)

	if len(l.logs) > l.capacity {
		l.logs = l.logs[:l.capacity]
	}
}

func (l *WrappedLogger) LastLogs() []MemoryLogDTO {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append([]MemoryLogDTO(nil), l.logs...)
}

func (l *WrappedLogger) Debug(args ...any) {
	l.logger.Debug(args...)
	l.log(logger.LogLevelDebug, fmt.Sprint(args...))
}

func (l *WrappedLogger) Debugln(args ...any) {
	l.logger.Debugln(args...)
	l.log(logger.LogLevelDebug, fmt.Sprintln(args...))
}

func (l *WrappedLogger) Debugf(template string, args ...any) {
	l.logger.Debugf(template, args...)
	l.log(logger.LogLevelDebug, fmt.Sprintf(template, args...))
}

func (l *WrappedLogger) Debugw(msg string, keysAndValues ...any) {
	l.logger.Debugw(msg, keysAndValues...)
	l.log(logger.LogLevelDebug, msg, keysAndValues...)
}

func (l *WrappedLogger) Info(args ...any) {
	l.logger.Info(args...)
	l.log(logger.LogLevelInfo, fmt.Sprint(args...))
}

func (l *WrappedLogger) Infoln(args ...any) {
	l.logger.Infoln(args...)
	l.log(logger.LogLevelInfo, fmt.Sprintln(args...))
}

func (l *WrappedLogger) Infof(template string, args ...any) {
	l.logger.Infof(template, args...)
	l.log(logger.LogLevelInfo, fmt.Sprintf(template, args...))
}

func (l *WrappedLogger) Infow(msg string, keysAndValues ...any) {
	l.logger.Infow(msg, keysAndValues...)
	l.log(logger.LogLevelInfo, msg, keysAndValues...)
}

func (l *WrappedLogger) Warn(args ...any) {
	l.logger.Warn(args...)
	l.log(logger.LogLevelWarn, fmt.Sprint(args...))
}

func (l *WrappedLogger) Warnln(args ...any) {
	l.logger.Warnln(args...)
	l.log(logger.LogLevelWarn, fmt.Sprintln(args...))
}

func (l *WrappedLogger) Warnf(template string, args ...any) {
	l.logger.Warnf(template, args...)
	l.log(logger.LogLevelWarn, fmt.Sprintf(template, args...))
}

func (l *WrappedLogger) Warnw(msg string, keysAndValues ...any) {
	l.logger.Warnw(msg, keysAndValues...)
	l.log(logger.LogLevelWarn, msg, keysAndValues...)
}

func (l *WrappedLogger) Error(args ...any) {
	l.logger.Error(args...)
	l.log(logger.LogLevelError, fmt.Sprint(args...))
}

func (l *WrappedLogger) Errorln(args ...any) {
	l.logger.Errorln(args...)
	l.log(logger.LogLevelError, fmt.Sprintln(args...))
}

func (l *WrappedLogger) Errorf(template string, args ...any) {
	l.logger.Errorf(template, args...)
	l.log(logger.LogLevelError, fmt.Sprintf(template, args...))
}

func (l *WrappedLogger) Errorw(msg string, keysAndValues ...any) {
	l.logger.Errorw(msg, keysAndValues...)
	l.log(logger.LogLevelError, msg, keysAndValues...)
}

func (l *WrappedLogger) Fatal(args ...any) {
	l.logger.Fatal(args...)
	l.log(logger.LogLevelFatal, fmt.Sprint(args...))
}

func (l *WrappedLogger) Fatalln(args ...any) {
	l.logger.Fatalln(args...)
	l.log(logger.LogLevelFatal, fmt.Sprintln(args...))
}

func (l *WrappedLogger) Fatalf(template string, args ...any) {
	l.logger.Fatalf(template, args...)
	l.log(logger.LogLevelFatal, fmt.Sprintf(template, args...))
}

func (l *WrappedLogger) Fatalw(msg string, keysAndValues ...any) {
	l.logger.Fatalw(msg, keysAndValues...)
	l.log(logger.LogLevelFatal, msg, keysAndValues...)
}

func (l *WrappedLogger) logToDB(level string, msg string, status string, processID uuid.UUID, typeSlug string) {
	if l.dbSyncer == nil {
		return
	}

	logDTO := LogDTO{
		TypeSlug:  typeSlug,
		Level:     level,
		Message:   msg,
		Status:    status,
		ProcessID: processID,
	}
	_, _ = l.dbSyncer.WriteLog(context.Background(), logDTO)
}

func (l *WrappedLogger) WithDBSyncer(dbSyncer *DBWriteSyncer) {
	l.dbSyncer = dbSyncer
}

func (l *WrappedLogger) Sugar() *zap.SugaredLogger {
	return l.sugaredLogger
}

func (l *WrappedLogger) Std() *log.Logger {
	return zap.NewStdLog(l.Desugar())
}

func (l *WrappedLogger) log(level logger.LogLevel, msg string, args ...any) {
	if safeLevel(level) < l.minLevel {
		return
	}
	var logParams *LogPrams
	if len(args) > 0 {
		if p, ok := args[0].(*LogPrams); ok {
			logParams = p
		}
	}
	if logParams != nil {
		l.logToDB(level.String(), msg, logParams.Status.String(), logParams.ProcessID, logParams.Slug)
	} else {
		l.addToMemory(level.String(), msg)
	}
}

func safeLevel(level logger.LogLevel) zapcore.Level {
	switch level {
	default:
		return zapcore.InfoLevel
	case logger.LogLevelDebug:
		return zapcore.DebugLevel
	case logger.LogLevelWarn:
		return zapcore.WarnLevel
	case logger.LogLevelError:
		return zapcore.ErrorLevel
	case logger.LogLevelPanic:
		return zapcore.PanicLevel
	case logger.LogLevelFatal:
		return zapcore.FatalLevel
	}
}
