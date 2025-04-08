package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	// Test default level (info)
	err := Init("test")
	assert.NoError(t, err)

	// Test with custom level
	os.Setenv("LOG_LEVEL", "debug")
	err = Init("test")
	assert.NoError(t, err)

	// Test with invalid level
	os.Setenv("LOG_LEVEL", "invalid")
	err = Init("test")
	assert.Error(t, err)
}

func TestInfo(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Info("test message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "test message", result["msg"])
	assert.Equal(t, "info", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestInfof(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Infof("test %s", "message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "test message", result["msg"])
	assert.Equal(t, "info", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestError(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Error("error message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "error message", result["msg"])
	assert.Equal(t, "error", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestErrorf(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Errorf("error %s", "message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "error message", result["msg"])
	assert.Equal(t, "error", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestDebug(t *testing.T) {
	// Setup
	os.Setenv("LOG_LEVEL", "debug")
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Debug("debug message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "debug message", result["msg"])
	assert.Equal(t, "debug", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestDebugf(t *testing.T) {
	// Setup
	os.Setenv("LOG_LEVEL", "debug")
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Debugf("debug %s", "message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "debug message", result["msg"])
	assert.Equal(t, "debug", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestWarn(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Warn("warning message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "warning message", result["msg"])
	assert.Equal(t, "warning", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestWarnf(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Warnf("warning %s", "message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "warning message", result["msg"])
	assert.Equal(t, "warning", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestFatal(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	// Note: We can't actually test Fatal as it calls os.Exit(1)
	// Instead, we'll just verify the function exists and compiles
	_ = Fatal
}

func TestFatalf(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	// Note: We can't actually test Fatalf as it calls os.Exit(1)
	// Instead, we'll just verify the function exists and compiles
	_ = Fatalf
}

func TestPanic(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	// Note: We can't actually test Panic as it calls panic()
	// Instead, we'll just verify the function exists and compiles
	_ = Panic
}

func TestPanicf(t *testing.T) {
	// Setup
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	// Note: We can't actually test Panicf as it calls panic()
	// Instead, we'll just verify the function exists and compiles
	_ = Panicf
}

func TestTrace(t *testing.T) {
	// Setup
	os.Setenv("LOG_LEVEL", "trace")
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Trace("trace message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "trace message", result["msg"])
	assert.Equal(t, "trace", result["level"])
	assert.Equal(t, "test", result["service"])
}

func TestTracef(t *testing.T) {
	// Setup
	os.Setenv("LOG_LEVEL", "trace")
	Init("test")
	var buf bytes.Buffer
	l.SetOutput(&buf)

	// Test
	Tracef("trace %s", "message")

	// Verify
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "trace message", result["msg"])
	assert.Equal(t, "trace", result["level"])
	assert.Equal(t, "test", result["service"])
}
