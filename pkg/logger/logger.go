package logger

import (
	"context"
	"log"

	"github.com/dv-net/mx/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type sugaredLogger = zap.SugaredLogger

type Logger interface {
	logger.ExtendedLogger
	WithDBSyncer(dbSyncer *DBWriteSyncer)
}

type WrappedLogger struct {
	logger   logger.Logger
	dbSyncer *DBWriteSyncer

	*sugaredLogger
}

func (l *WrappedLogger) Sugar() *zap.SugaredLogger {
	return l.sugaredLogger
}

func (l *WrappedLogger) Std() *log.Logger {
	return zap.NewStdLog(l.Desugar())
}

type LogPrams struct {
	Status    LogStatus
	Slug      string
	ProcessID uuid.UUID
}

var _ Logger = (*WrappedLogger)(nil)

func New(appVersion string, conf logger.Config) Logger {
	log := logger.NewExtended(
		logger.WithLogFormat(logger.LoggerFormatJSON),
		logger.WithAppVersion(appVersion),
		logger.WithConfig(conf),
	)

	return &WrappedLogger{logger: log, sugaredLogger: log.Sugar()}
}

func (l *WrappedLogger) Warnf(template string, args ...any) {
	l.logger.Warnf(template, args...)
	var logParams *LogPrams
	if len(args) > 0 {
		if p, ok := args[0].(*LogPrams); ok {
			logParams = p
		}
	}
	if logParams != nil {
		l.logToDB("WARN", template, logParams.Status.String(), logParams.ProcessID, logParams.Slug)
	}
}

func (l *WrappedLogger) Warnw(msg string, keysAndValues ...any) {
	l.logger.Warnw(msg, keysAndValues...)
	var logParams *LogPrams
	if len(keysAndValues) > 0 {
		if p, ok := keysAndValues[0].(*LogPrams); ok {
			logParams = p
		}
	}
	if logParams != nil {
		l.logToDB("WARN", msg, logParams.Status.String(), logParams.ProcessID, logParams.Slug)
	}
}

func (l *WrappedLogger) Infow(msg string, params ...any) {
	l.logger.Infow(msg, params...)
	var logParams *LogPrams
	if len(params) > 0 {
		if p, ok := params[0].(*LogPrams); ok {
			logParams = p
		}
	}
	if logParams != nil {
		l.logToDB("INFO", msg, logParams.Status.String(), logParams.ProcessID, logParams.Slug)
	}
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
