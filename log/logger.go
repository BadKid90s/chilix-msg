package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
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

// logWithCaller 记录日志并包含文件名和行号
func (l *DefaultLogger) logWithCaller(level, format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		filename := filepath.Base(file)
		l.logger.Printf(fmt.Sprintf("[%s] %s:%d %s", level, filename, line, format), v...)
	} else {
		l.logger.Printf(fmt.Sprintf("[%s] %s", level, format), v...)
	}
}

// Infof 记录信息级别日志
func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	l.logWithCaller("INFO", format, v...)
}

// Errorf 记录错误级别日志
func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	l.logWithCaller("ERROR", format, v...)
}

// Debugf 记录调试级别日志
func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	l.logWithCaller("DEBUG", format, v...)
}

// Warnf 记录警告级别日志
func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	l.logWithCaller("WARN", format, v...)
}

// Fatalf 记录致命错误日志
func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		filename := filepath.Base(file)
		l.logger.Fatalf(fmt.Sprintf("[FATAL] %s:%d %s", filename, line, format), v...)
	} else {
		l.logger.Fatalf(fmt.Sprintf("[FATAL] %s", format), v...)
	}
}

// GetLogger 获取底层的logger实例
func (l *DefaultLogger) GetLogger() *log.Logger {
	return l.logger
}

// 全局默认日志器
var defaultLogger Logger = NewDefaultLogger()

// SetDefaultLogger 设置全局默认日志器
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// GetDefaultLogger 获取全局默认日志器
func GetDefaultLogger() Logger {
	return defaultLogger
}

// 包级别的日志函数，自动显示文件名和行号

// logWithCallerForGlobal 全局函数使用的日志记录函数
func logWithCallerForGlobal(level, format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2) // 跳过当前函数和包级别函数
	if ok {
		filename := filepath.Base(file)
		defaultLogger.(*DefaultLogger).logger.Printf(fmt.Sprintf("[%s] %s:%d %s", level, filename, line, format), v...)
	} else {
		defaultLogger.(*DefaultLogger).logger.Printf(fmt.Sprintf("[%s] %s", level, format), v...)
	}
}

// Infof 记录信息级别日志，自动包含文件名和行号
func Infof(format string, v ...interface{}) {
	logWithCallerForGlobal("INFO", format, v...)
}

// Errorf 记录错误级别日志，自动包含文件名和行号
func Errorf(format string, v ...interface{}) {
	logWithCallerForGlobal("ERROR", format, v...)
}

// Debugf 记录调试级别日志，自动包含文件名和行号
func Debugf(format string, v ...interface{}) {
	logWithCallerForGlobal("DEBUG", format, v...)
}

// Warnf 记录警告级别日志，自动包含文件名和行号
func Warnf(format string, v ...interface{}) {
	logWithCallerForGlobal("WARN", format, v...)
}

// Fatalf 记录致命错误日志，自动包含文件名和行号
func Fatalf(format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		filename := filepath.Base(file)
		defaultLogger.(*DefaultLogger).logger.Fatalf(fmt.Sprintf("[FATAL] %s:%d %s", filename, line, format), v...)
	} else {
		defaultLogger.(*DefaultLogger).logger.Fatalf(fmt.Sprintf("[FATAL] %s", format), v...)
	}
}
