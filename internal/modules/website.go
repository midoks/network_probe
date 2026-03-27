package modules

import (
	"fmt"
	"net/http"
	"time"
)

// WebsiteTestConfig 表示网站测试配置
type WebsiteTestConfig struct {
	URL             string
	Method          string
	Timeout         time.Duration
	FollowRedirects bool
}

// NewWebsiteTestConfig 创建一个新的网站测试配置
func NewWebsiteTestConfig() *WebsiteTestConfig {
	return &WebsiteTestConfig{
		Method:          "GET",
		Timeout:         30 * time.Second,
		FollowRedirects: true,
	}
}

// WebsiteTestResult 表示网站测试结果
type WebsiteTestResult struct {
	URL             string
	Status          int
	StatusText      string
	ResponseTime    float64
	ContentLength   int64
	ContentType     string
	RedirectURL     string
	Error           string
}

// WebsiteTestService 表示网站测试服务
type WebsiteTestService struct{}

// NewWebsiteTestService 创建一个新的网站测试服务
func NewWebsiteTestService() *WebsiteTestService {
	return &WebsiteTestService{}
}

// TestWebsite 测试网站
func (s *WebsiteTestService) TestWebsite(config *WebsiteTestConfig) (*WebsiteTestResult, error) {
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: config.Timeout,
	}

	if !config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest(config.Method, config.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 发送请求并记录时间
	start := time.Now()
	resp, err := client.Do(req)
	responseTime := time.Since(start)

	result := &WebsiteTestResult{
		URL:          config.URL,
		ResponseTime: responseTime.Seconds() * 1000,
	}

	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	// 填充响应信息
	result.Status = resp.StatusCode
	result.StatusText = resp.Status
	result.ContentLength = resp.ContentLength
	result.ContentType = resp.Header.Get("Content-Type")

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		result.RedirectURL = resp.Header.Get("Location")
	}

	return result, nil
}
