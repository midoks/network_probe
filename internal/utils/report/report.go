package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"network-probe/internal/config"
	"network-probe/internal/utils/system"
	"network-probe/internal/version"
)

// PingRequest 表示 ping 请求
type PingRequest struct {
	Host    string `json:"host"`
	Count   int    `json:"count"`
	Timeout int    `json:"timeout"`
}

// PingReportData 表示 ping 上报数据
type PingReportData struct {
	Request PingRequest `json:"request"`
	Result  interface{} `json:"result"`
}

// TcpingRequest 表示 tcping 请求
type TcpingRequest struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Count   int    `json:"count"`
	Timeout int    `json:"timeout"`
}

// TcpingReportData 表示 tcping 上报数据
type TcpingReportData struct {
	Request TcpingRequest `json:"request"`
	Result  interface{}   `json:"result"`
}

// WebsiteRequest 表示 website 请求
type WebsiteRequest struct {
	URL             string `json:"url"`
	Method          string `json:"method"`
	Timeout         int    `json:"timeout"`
	FollowRedirects bool   `json:"follow_redirects"`
}

// WebsiteReportData 表示 website 上报数据
type WebsiteReportData struct {
	Request WebsiteRequest `json:"request"`
	Result  interface{}    `json:"result"`
}

// TracerouteRequest 表示 traceroute 请求
type TracerouteRequest struct {
	Host     string `json:"host"`
	MaxHops  int    `json:"max_hops"`
	Protocol string `json:"protocol"`
}

// TracerouteReportData 表示 traceroute 上报数据
type TracerouteReportData struct {
	Request TracerouteRequest `json:"request"`
	Result  interface{}       `json:"result"`
}

// DnsRequest 表示 dns 请求
type DnsRequest struct {
	Domain     string `json:"domain"`
	QueryType  string `json:"query_type"`
	Nameserver string `json:"nameserver"`
}

// DnsReportData 表示 dns 上报数据
type DnsReportData struct {
	Request DnsRequest  `json:"request"`
	Result  interface{} `json:"result"`
}

// MtrRequest 表示 mtr 请求
type MtrRequest struct {
	Host     string `json:"host"`
	MaxHops  int    `json:"max_hops"`
	Count    int    `json:"count"`
	Interval int    `json:"interval"`
}

// MtrReportData 表示 mtr 上报数据
type MtrReportData struct {
	Request MtrRequest  `json:"request"`
	Result  interface{} `json:"result"`
}

// StartupRequest 表示启动请求
type StartupRequest struct {
	Address string `json:"address"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// StartupReportData 表示启动上报数据
type StartupReportData struct {
	Msg string `json:"msg"`
}

// WebSocketConnectReportData 表示 WebSocket 连接上报数据
type WebSocketConnectReportData struct {
	Status string `json:"status"`
}

// SystemInfoReportData 表示系统信息上报数据
type SystemInfoReportData struct {
	Result interface{} `json:"result"`
}

// 全局 channel 用于控制定时上传
type uploadTask struct {
	reportType ReportType
	subType    SubType
	data       interface{}
}

var (
	uploadChan chan uploadTask
	stopChan   chan struct{}
	isRunning  bool
)

func init() {
	// 初始化 channel
	uploadChan = make(chan uploadTask, 100)
	stopChan = make(chan struct{})
	isRunning = true

	// 启动定时上传 goroutine
	go uploadWorker()
}

// uploadWorker 定时上传工作器
func uploadWorker() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 定时触发上传日志
			fmt.Println("[LOG]定时上传任务触发")
		case task := <-uploadChan:
			// 处理实时上传任务
			if err := Report(task.reportType, task.subType, task.data); err != nil {
				fmt.Printf("[LOG]upload task failed: %v\n", err)
			}
		case <-stopChan:
			// 停止上传
			fmt.Println("[LOG]upload worker stopped")
			return
		}
	}
}

// StopUploadWorker 停止上传工作器
func StopUploadWorker() {
	if isRunning {
		close(stopChan)
		isRunning = false
	}
}

// QueueUpload 将上传任务加入队列
func QueueUpload(reportType ReportType, subType SubType, data interface{}) {
	select {
	case uploadChan <- uploadTask{
		reportType: reportType,
		subType:    subType,
		data:       data,
	}:
		// 任务已加入队列
	default:
		// 队列已满，直接上报
		fmt.Println("[LOG]upload queue is full, reporting directly")
		if err := Report(reportType, subType, data); err != nil {
			fmt.Printf("[LOG]direct upload failed: %v\n", err)
		}
	}
}

// Report 上报数据（同步）
func Report(reportType ReportType, subType SubType, data interface{}) error {
	// 序列化数据
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// 准备上报数据
	reportDataReady := ReportData{
		Type:      reportType,
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
		Data:      string(dataBytes),
	}

	// 序列化上报数据
	reportData, err := json.Marshal(reportDataReady)
	if err != nil {
		return fmt.Errorf("failed to marshal report data: %v", err)
	}

	return ReportBytes(reportData)
}

// ReportBytes 上报数据（字节数组）
func ReportBytes(data []byte) error {
	cfg, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if len(cfg.ReportEndpoints) == 0 {
		return fmt.Errorf("no report endpoints configured")
	}

	// 上报到每个端点（同步）
	var lastError error
	for _, endpoint := range cfg.ReportEndpoints {

		url := endpoint + "/api/logs"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			lastError = fmt.Errorf("failed to create request to %s: %v", endpoint, err)
			fmt.Printf("Failed to create request to %s: %v\n", url, err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Node-ID", cfg.NodeID)
		req.Header.Set("X-Secret", cfg.Secret)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastError = fmt.Errorf("failed to report to %s: %v", url, err)
			fmt.Printf("Failed to report to %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()
		fmt.Printf("Reported to %s successfully: %v\n", url, resp.Status)
	}

	return lastError
}

// 上报节点信息记录
func NodeInfo(tag, description string) error {

	// 准备上报数据
	reportDataReady := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	fmt.Println("[" + tag + "]" + description)

	// 设置节点日志数据
	err := reportDataReady.SetNodeLogsData(ReportNodeLogs{
		Tag:         tag,
		Level:       "info",
		Description: description,
		CreateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to set node logs data: %v", err)
	}

	// 序列化数据
	reportData, err := json.Marshal(reportDataReady)
	if err != nil {
		return fmt.Errorf("failed to marshal report info data: %v", err)
	}
	return ReportBytes(reportData)
}

// 上报节点警告记录
func NodeWarn(tag, description string) error {

	// 准备上报数据
	reportDataReady := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	fmt.Println("[" + tag + "]" + description)

	// 设置节点日志数据
	err := reportDataReady.SetNodeLogsData(ReportNodeLogs{
		Tag:         tag,
		Level:       "warning",
		Description: description,
		CreateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to set node warning logs data: %v", err)
	}

	// 序列化数据
	reportData, err := json.Marshal(reportDataReady)
	if err != nil {
		return fmt.Errorf("failed to marshal report warning data: %v", err)
	}
	return ReportBytes(reportData)
}

// 上报节点错误记录
func NodeError(tag, description string) error {

	// 准备上报数据
	reportDataReady := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	fmt.Println("[" + tag + "]" + description)

	// 设置节点日志数据
	err := reportDataReady.SetNodeLogsData(ReportNodeLogs{
		Tag:         tag,
		Level:       "error",
		Description: description,
		CreateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to set node error logs data: %v", err)
	}

	// 序列化数据
	reportData, err := json.Marshal(reportDataReady)
	if err != nil {
		return fmt.Errorf("failed to marshal report info error data: %v", err)
	}
	return ReportBytes(reportData)
}

// 上报节点成功记录
func NodeSuccess(tag, description string) error {
	// 准备上报数据
	reportDataReady := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	fmt.Println("[" + tag + "]" + description)

	// 设置节点日志数据
	err := reportDataReady.SetNodeLogsData(ReportNodeLogs{
		Tag:         tag,
		Level:       "success",
		Description: description,
		CreateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to set node error logs data: %v", err)
	}

	// 序列化数据
	reportData, err := json.Marshal(reportDataReady)
	if err != nil {
		return fmt.Errorf("failed to marshal report info error data: %v", err)
	}
	return ReportBytes(reportData)
}

// 节点 cpu/mem/disk 信息
func NodeItem(item string, value interface{}) error {
	report := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	item_value, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal report node item value: %v", err)
	}

	fmt.Println("item", item, string(item_value))
	report.SetNodeItemData(ReportNodeItem{
		Item:  item,
		Value: string(item_value),
	})

	report_data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report node item: %v", err)
	}
	return ReportBytes(report_data)
}

func ReportSystemInfo() error {
	fmt.Println("定时器触发，开始获取系统信息")
	// 获取系统信息
	_, err := system.GetSystemInfo()
	if err != nil {
		fmt.Printf("获取系统信息失败: %v\n", err)
		return err
	}
	fmt.Println("获取系统信息成功，开始上报")

	// 准备上报数据
	reportDataReady := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	reportDataReady.SetSysInfoData(ReportSysInfo{
		CreateTime: time.Now().Unix(),
	})

	reportData, err := json.Marshal(reportDataReady)
	if err != nil {
		return fmt.Errorf("failed to marshal report sysinfo data: %v", err)
	}
	return ReportBytes(reportData)
}

func ReportRequest(data interface{}) error {
	// 准备上报数据
	ready := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}
	ready.SetRequestLogsData(data)
	report_data, err := json.Marshal(ready)
	if err != nil {
		return fmt.Errorf("failed to marshal report request cmd data: %v", err)
	}
	return ReportBytes(report_data)
}

// ReportErrorLog 上报错误日志
func ReportErrorLog(entry interface{}) error {
	return Report(ReportTypeSystem, "error_log", map[string]interface{}{
		"error": entry,
	})
}

// ReportCrashLog 上报崩溃日志
func ReportCrashLog(entry interface{}) error {
	return Report(ReportTypeSystem, "crash_log", map[string]interface{}{
		"crash": entry,
	})
}
