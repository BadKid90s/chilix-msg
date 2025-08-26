package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试创建默认日志记录器
func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.GetLogger())
}

// 测试Info级别日志
func TestDefaultLogger_Infof(t *testing.T) {
	var buf bytes.Buffer

	// 替换底层logger的输出
	testLogger := NewDefaultLogger()
	testLogger.GetLogger().SetOutput(&buf)

	testLogger.Infof("Test info message: %s", "test")
	output := buf.String()
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "logger_test.go:")
	assert.Contains(t, output, "Test info message: test")
}

// 测试Error级别日志
func TestDefaultLogger_Errorf(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewDefaultLogger()
	testLogger.GetLogger().SetOutput(&buf)

	testLogger.Errorf("Test error message: %s", "test")
	output := buf.String()
	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "logger_test.go:")
	assert.Contains(t, output, "Test error message: test")
}

// 测试Debug级别日志
func TestDefaultLogger_Debugf(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewDefaultLogger()
	testLogger.GetLogger().SetOutput(&buf)

	testLogger.Debugf("Test debug message: %s", "test")
	output := buf.String()
	assert.Contains(t, output, "[DEBUG]")
	assert.Contains(t, output, "logger_test.go:")
	assert.Contains(t, output, "Test debug message: test")
}

// 测试Warn级别日志
func TestDefaultLogger_Warnf(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewDefaultLogger()
	testLogger.GetLogger().SetOutput(&buf)

	testLogger.Warnf("Test warn message: %s", "test")
	output := buf.String()
	assert.Contains(t, output, "[WARN]")
	assert.Contains(t, output, "logger_test.go:")
	assert.Contains(t, output, "Test warn message: test")
}

// 测试包级别日志函数
func TestPackageLevelFunctions(t *testing.T) {
	// 保存原始的默认日志器
	originalLogger := defaultLogger
	defer SetDefaultLogger(originalLogger)

	// 创建测试用的日志器
	var buf bytes.Buffer
	testLogger := NewDefaultLogger()
	testLogger.GetLogger().SetOutput(&buf)
	SetDefaultLogger(testLogger)

	// 测试包级别函数
	Infof("信息消息: %s", "test")
	Errorf("错误消息: %d", 404)
	Debugf("调试消息")
	Warnf("警告消息")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// 验证日志级别和文件名
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "[DEBUG]")
	assert.Contains(t, output, "[WARN]")

	// 验证文件名和行号信息
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			assert.Contains(t, line, "logger_test.go:")
		}
	}

	// 验证消息内容
	assert.Contains(t, output, "信息消息: test")
	assert.Contains(t, output, "错误消息: 404")
	assert.Contains(t, output, "调试消息")
	assert.Contains(t, output, "警告消息")
}
