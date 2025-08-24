package log

import (
	"log"
	"os"
)

// Logger 定义日志记录器接口
type Logger interface {
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

// DefaultLogger 默认日志记录器实现
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger 创建默认日志记录器
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Infof 记录信息级别日志
func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	l.logger.Printf("[INFO] "+format, v...)
}

// Errorf 记录错误级别日志
func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	l.logger.Printf("[ERROR] "+format, v...)
}

// Debugf 记录调试级别日志
func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	l.logger.Printf("[DEBUG] "+format, v...)
}

// Warnf 记录警告级别日志
func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	l.logger.Printf("[WARN] "+format, v...)
}

func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf("[FATAL] "+format, v...)
}
