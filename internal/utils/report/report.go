package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"network-probe/internal/config"
	"network-probe/internal/version"
)

// 全局 用于控制定时上传
type uploadTask struct {
	data ReportData
}

var (
	uploadChan chan uploadTask
)

func init() {
	// 初始化
	uploadChan = make(chan uploadTask, 64)

	// 启动定时上传
	go uploadWorker()
}

// 定时上传工作器
func uploadWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		uploadLogs()
	}

}

func uploadLogs() error {
Loop:
	for {
		select {
		case task := <-uploadChan:
			dataBytes, err := json.Marshal(task.data)
			if err != nil {
				fmt.Printf("[LOG]failed to marshal report data: %v\n", err)
				continue
			}
			// fmt.Println("data;", string(dataBytes))
			if err := ReportBytes(dataBytes); err != nil {
				fmt.Printf("[LOG]upload task failed: %v\n", err)
			}

			// task_cap := len(uploadChan)
			// fmt.Println("task cap:", task_cap)
		default:
			//loop end
			break Loop
		}
	}
	return nil
}

// 上报数据
func Report(data interface{}) error {
	ready := ReportData{
		Type:      ReportTypeNode,
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal report data: %v", err)
	}
	ready.Data = string(dataBytes)

	report_data, err := json.Marshal(ready)
	if err != nil {
		return fmt.Errorf("failed to marshal report data: %v", err)
	}
	return ReportBytes(report_data)
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
			fmt.Printf("failed to report to %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()
		fmt.Printf("reported to %s successfully: %v\n", url, resp.Status)
	}

	return lastError
}

// 上报节点信息记录
func NodeInfo(tag, description string) error {
	ready := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}
	err := ready.SetNodeLogsData(ReportNodeLogs{
		Tag:         tag,
		Level:       "info",
		Description: description,
		CreateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to set node logs data: %v", err)
	}

	select {
	case uploadChan <- uploadTask{
		data: ready,
	}:
	default:
	}
	return nil
}

// 上报节点警告记录
func NodeWarn(tag, description string) error {
	ready := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}
	err := ready.SetNodeLogsData(ReportNodeLogs{
		Tag:         tag,
		Level:       "warning",
		Description: description,
		CreateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to set node warning logs data: %v", err)
	}

	select {
	case uploadChan <- uploadTask{
		data: ready,
	}:
	default:
	}
	return nil
}

// 上报节点错误记录
func NodeError(tag, description string) error {
	ready := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	err := ready.SetNodeLogsData(ReportNodeLogs{
		Tag:         tag,
		Level:       "error",
		Description: description,
		CreateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to set node error logs data: %v", err)
	}

	select {
	case uploadChan <- uploadTask{
		data: ready,
	}:
	default:
	}
	return nil
}

// 上报节点成功记录
func NodeSuccess(tag, description string) error {
	ready := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}
	// fmt.Println("[" + tag + "]" + description)
	err := ready.SetNodeLogsData(ReportNodeLogs{
		Tag:         tag,
		Level:       "success",
		Description: description,
		CreateTime:  time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to set node error logs data: %v", err)
	}

	select {
	case uploadChan <- uploadTask{
		data: ready,
	}:
	default:
	}
	return nil
}

// 节点 cpu/mem/disk/sysinfo 信息
func NodeItem(item string, value interface{}) error {
	ready := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	item_value, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal report node item value: %v", err)
	}
	ready.SetNodeItemData(ReportNodeItem{
		Item:  item,
		Value: string(item_value),
	})

	select {
	case uploadChan <- uploadTask{
		data: ready,
	}:
	default:
	}
	return nil
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
	ready := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	data := map[string]interface{}{
		"error": entry,
	}
	report_data, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal report error log data: %v", err)
	}
	ready.Data = string(report_data)

	select {
	case uploadChan <- uploadTask{
		data: ready,
	}:
	default:
	}
	return nil
}

// ReportCrashLog 上报崩溃日志
func ReportCrashLog(entry interface{}) error {
	ready := ReportData{
		Timestamp: time.Now().Unix(),
		Version:   version.Version,
	}

	data := map[string]interface{}{
		"crash": entry,
	}
	report_data, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal report crash log data: %v", err)
	}
	ready.Data = string(report_data)

	select {
	case uploadChan <- uploadTask{
		data: ready,
	}:
	default:
	}
	return nil
}
