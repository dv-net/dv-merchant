package logger

import (
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
	logger   logger.ExtendedLogger
	dbSyncer *DBWriteSyncer

	*sugaredLogger
}

func (l *WrappedLogger) Sugar() *zap.SugaredLogger {
	return l.sugaredLogger
}

func (l *WrappedLogger) Std() *log.Logger {
	return zap.NewStdLog(l.Desugar())
}

func (l WrappedLogger) LastLogs() []logger.MemoryLog {
	return l.logger.LastLogs()
}

func (l *WrappedLogger) WithDBSyncer(dbSyncer *DBWriteSyncer) {
	l.dbSyncer = dbSyncer
}

type LogPrams struct {
	Status    LogStatus
	Slug      string
	ProcessID uuid.UUID
}

var _ Logger = (*WrappedLogger)(nil)

func New(appVersion string, conf logger.Config) Logger {
	l := logger.NewExtended(
		logger.WithLogFormat(conf.Format),
		logger.WithAppVersion(appVersion),
		logger.WithConfig(conf),
		logger.WithMemoryBuffer(LogBufferSize),
	)

	return &WrappedLogger{
		logger:        l,
		sugaredLogger: l.Sugar(),
	}
}
