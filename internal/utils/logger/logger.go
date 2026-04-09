package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"network-probe/internal/utils/report"
)

const (
	// 日志文件路径
	ErrorLogFile = "data/error.log"
	CrashLogFile = "data/crash.log"
)

// ErrorLogEntry 表示错误日志条目
type ErrorLogEntry struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}

// CrashLogEntry 表示崩溃日志条目
type CrashLogEntry struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Stack     string `json:"stack"`
}

var (
	errorLogMutex sync.Mutex
	crashLogMutex sync.Mutex
)

var stdErrFileHandler *os.File

func RewriteStderrFile() error {
	file, err := os.OpenFile(CrashLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	stdErrFileHandler = file

	// 分析panic
	data, err := os.ReadFile(CrashLogFile)
	if err == nil {
		var index = bytes.Index(data, []byte("panic:"))
		if index >= 0 {
			report.NodeError("NODE", "系统错误，请上报给开发者: "+string(data[index:]))
		}
	}

	if err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		return err
	}

	runtime.SetFinalizer(stdErrFileHandler, func(fd *os.File) {
		fd.Close()
	})
	return nil
}

// LogError 记录错误信息
func LogError(errType, message, level string) {
	errorLogMutex.Lock()
	defer errorLogMutex.Unlock()

	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(ErrorLogFile), 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}

	// 创建日志条目
	entry := ErrorLogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      errType,
		Message:   message,
		Level:     level,
	}

	// 序列化日志条目
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal error log entry: %v\n", err)
		return
	}

	// 写入日志文件
	file, err := os.OpenFile(ErrorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open error log file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.Write(append(data, '\n')); err != nil {
		fmt.Printf("Failed to write error log: %v\n", err)
		return
	}
}

// LogCrash 记录崩溃信息
func LogCrash(errType, message, stack string) {
	crashLogMutex.Lock()
	defer crashLogMutex.Unlock()

	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(CrashLogFile), 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}

	// 创建日志条目
	entry := CrashLogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      errType,
		Message:   message,
		Stack:     stack,
	}

	// 序列化日志条目
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal crash log entry: %v\n", err)
		return
	}

	// 写入日志文件
	file, err := os.OpenFile(CrashLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open crash log file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.Write(append(data, '\n')); err != nil {
		fmt.Printf("Failed to write crash log: %v\n", err)
		return
	}
}

// ReadErrorLogs 读取错误日志
func ReadErrorLogs() ([]ErrorLogEntry, error) {
	errorLogMutex.Lock()
	defer errorLogMutex.Unlock()

	// 检查日志文件是否存在
	if _, err := os.Stat(ErrorLogFile); os.IsNotExist(err) {
		return nil, nil
	}

	// 读取日志文件
	data, err := os.ReadFile(ErrorLogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read error log file: %v", err)
	}

	// 解析日志条目
	var entries []ErrorLogEntry
	lines := splitLines(data)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var entry ErrorLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			fmt.Printf("Failed to unmarshal error log entry: %v\n", err)
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// ReadCrashLogs 读取崩溃日志
func ReadCrashLogs() ([]CrashLogEntry, error) {
	crashLogMutex.Lock()
	defer crashLogMutex.Unlock()

	// 检查日志文件是否存在
	if _, err := os.Stat(CrashLogFile); os.IsNotExist(err) {
		return nil, nil
	}

	// 读取日志文件
	data, err := os.ReadFile(CrashLogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read crash log file: %v", err)
	}

	// 解析日志条目
	var entries []CrashLogEntry
	lines := splitLines(data)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var entry CrashLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			fmt.Printf("Failed to unmarshal crash log entry: %v\n", err)
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// ClearErrorLogs 清除错误日志
func ClearErrorLogs() error {
	errorLogMutex.Lock()
	defer errorLogMutex.Unlock()

	if err := os.Remove(ErrorLogFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear error log file: %v", err)
	}

	return nil
}

// ClearCrashLogs 清除崩溃日志
func ClearCrashLogs() error {
	crashLogMutex.Lock()
	defer crashLogMutex.Unlock()

	if err := os.Remove(CrashLogFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear crash log file: %v", err)
	}

	return nil
}

// ReportErrorLogs 上报错误日志
func ReportErrorLogs() error {
	entries, err := ReadErrorLogs()
	if err != nil {
		return fmt.Errorf("failed to read error logs: %v", err)
	}

	if len(entries) == 0 {
		return nil
	}

	// 上报每个错误日志条目
	for _, entry := range entries {
		if err := report.ReportErrorLog(entry); err != nil {
			fmt.Printf("Failed to report error log: %v\n", err)
			continue
		}
	}

	// 清除已上报的错误日志
	return ClearErrorLogs()
}

// splitLines 分割数据为行
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
