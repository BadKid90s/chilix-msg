package log

import (
	"bytes"
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
	assert.Contains(t, output, "[INFO] Test info message: test")
}

// 测试Error级别日志
func TestDefaultLogger_Errorf(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewDefaultLogger()
	testLogger.GetLogger().SetOutput(&buf)

	testLogger.Errorf("Test error message: %s", "test")
	output := buf.String()
	assert.Contains(t, output, "[ERROR] Test error message: test")
}

// 测试Debug级别日志
func TestDefaultLogger_Debugf(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewDefaultLogger()
	testLogger.GetLogger().SetOutput(&buf)

	testLogger.Debugf("Test debug message: %s", "test")
	output := buf.String()
	assert.Contains(t, output, "[DEBUG] Test debug message: test")
}

// 测试Warn级别日志
func TestDefaultLogger_Warnf(t *testing.T) {
	var buf bytes.Buffer
	testLogger := NewDefaultLogger()
	testLogger.GetLogger().SetOutput(&buf)

	testLogger.Warnf("Test warn message: %s", "test")
	output := buf.String()
	assert.Contains(t, output, "[WARN] Test warn message: test")
}
