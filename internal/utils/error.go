package utils

import (
	"errors"
	"fmt"
)

// NetworkError 表示网络相关的错误
type NetworkError struct {
	Type    string
	Message string
}

// Error 实现 error 接口
func (e *NetworkError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewNetworkError 创建一个新的网络错误
func NewNetworkError(errType string, message string) *NetworkError {
	return &NetworkError{
		Type:    errType,
		Message: message,
	}
}

// NewIoError 创建一个 IO 错误
func NewIoError(err error) *NetworkError {
	return &NetworkError{
		Type:    "IO error",
		Message: err.Error(),
	}
}

// NewDnsError 创建一个 DNS 错误
func NewDnsError(message string) *NetworkError {
	return &NetworkError{
		Type:    "DNS resolution failed",
		Message: message,
	}
}

// NewPingError 创建一个 Ping 错误
func NewPingError(message string) *NetworkError {
	return &NetworkError{
		Type:    "Ping failed",
		Message: message,
	}
}

// NewTcpError 创建一个 TCP 错误
func NewTcpError(message string) *NetworkError {
	return &NetworkError{
		Type:    "TCP connection failed",
		Message: message,
	}
}

// NewHttpError 创建一个 HTTP 错误
func NewHttpError(message string) *NetworkError {
	return &NetworkError{
		Type:    "HTTP request failed",
		Message: message,
	}
}

// NewTracerouteError 创建一个 Traceroute 错误
func NewTracerouteError(message string) *NetworkError {
	return &NetworkError{
		Type:    "Traceroute failed",
		Message: message,
	}
}

// NewTimeoutError 创建一个超时错误
func NewTimeoutError(message string) *NetworkError {
	return &NetworkError{
		Type:    "Timeout",
		Message: message,
	}
}

// NewInvalidInputError 创建一个无效输入错误
func NewInvalidInputError(message string) *NetworkError {
	return &NetworkError{
		Type:    "Invalid input",
		Message: message,
	}
}

// NewPermissionError 创建一个权限错误
func NewPermissionError(message string) *NetworkError {
	return &NetworkError{
		Type:    "Permission denied",
		Message: message,
	}
}

// NewOtherError 创建一个其他错误
func NewOtherError(err error) *NetworkError {
	return &NetworkError{
		Type:    "Other error",
		Message: err.Error(),
	}
}

// IsNetworkError 检查错误是否为 NetworkError
func IsNetworkError(err error) (*NetworkError, bool) {
	var networkErr *NetworkError
	if errors.As(err, &networkErr) {
		return networkErr, true
	}
	return nil, false
}
