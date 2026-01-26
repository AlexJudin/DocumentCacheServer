package temporal

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/log"
)

// LogrusLogger реализует интерфейс log.Logger из Temporal SDK
type LogrusLogger struct {
	logger *logrus.Entry
}

// NewLogrusLogger создает новый экземпляр логгера
func NewLogrusLogger(logrusLogger *logrus.Logger) *LogrusLogger {
	return &LogrusLogger{
		logger: logrus.NewEntry(logrusLogger),
	}
}

// Debug логирует сообщение на уровне Debug
func (l *LogrusLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.WithFields(toFields(keyvals...)).Debug(msg)
}

// Info логирует сообщение на уровне Info
func (l *LogrusLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.WithFields(toFields(keyvals...)).Info(msg)
}

// Warn логирует сообщение на уровне Warn
func (l *LogrusLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.WithFields(toFields(keyvals...)).Warn(msg)
}

// Error логирует сообщение на уровне Error
func (l *LogrusLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.WithFields(toFields(keyvals...)).Error(msg)
}

// With добавляет поля к логгеру
func (l *LogrusLogger) With(keyvals ...interface{}) log.Logger {
	return &LogrusLogger{
		logger: l.logger.WithFields(toFields(keyvals...)),
	}
}

// toFields преобразует пары ключ-значение в fields для logrus
func toFields(keyvals ...interface{}) logrus.Fields {
	fields := make(logrus.Fields)

	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := fmt.Sprintf("%v", keyvals[i])
			fields[key] = keyvals[i+1]
		}
	}

	return fields
}
