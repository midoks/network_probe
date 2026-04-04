package report

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"network-probe/internal/config"
	"network-probe/internal/version"

	"github.com/shirou/gopsutil/v3/cpu"
)

type NodeStatusExecutor struct {
	isFirstTime     bool
	lastUpdatedTime time.Time

	cpuLogicalCount  int
	cpuPhysicalCount int

	// 流量统计
	lastUDPInDatagrams  int64
	lastUDPOutDatagrams int64

	ticker *time.Ticker
}

type NodeStatus struct {
	BuildVersion string `json:"build_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	ExePath      string `json:"exe_path"`
	Hostname     string `json:"hostname"`
	UpdatedAt    int64  `json:"updated_at"`
	Timestamp    int64  `json:"timestamp"`

	//cpu
	CPUUsage         float64 `json:"cpu_usage"`
	CPULogicalCount  int     `json:"cpu_logical_count"`
	CPUPhysicalCount int     `json:"cpu_physical_count"`

	IsActive bool   `json:"is_active"`
	Error    string `json:"error"`
}

func NewNodeStatusExecutor() *NodeStatusExecutor {
	return &NodeStatusExecutor{
		ticker: time.NewTicker(30 * time.Second),

		lastUDPInDatagrams:  -1,
		lastUDPOutDatagrams: -1,
	}
}

func (this *NodeStatusExecutor) Run() {
	this.isFirstTime = true
	this.lastUpdatedTime = time.Now()
	this.update()

	for range this.ticker.C {
		this.isFirstTime = false
		this.update()
	}
}

func (this *NodeStatusExecutor) update() {
	// 加载配置
	_, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// 构建节点状态
	status := &NodeStatus{
		BuildVersion: version.Version,
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		IsActive:     true,
		UpdatedAt:    time.Now().Unix(),
		Timestamp:    time.Now().Unix(),
	}

	// 获取可执行文件路径
	status.ExePath, _ = os.Executable()

	// 获取主机名
	hostname, _ := os.Hostname()
	status.Hostname = hostname

	// cpu
	this.updateCPU(status)

	// 序列化数据
	nodeData, err := json.Marshal(status)
	if err != nil {
		fmt.Printf("Failed to marshal node status: %v\n", err)
		return
	}

	fmt.Println("nodeData:", string(nodeData))

	// 上报数据
	if err := ReportBytes(nodeData); err != nil {
		fmt.Printf("failed to report node status: %v\n", err)
		return
	}

	// 修改更新时间
	this.lastUpdatedTime = time.Now()

	fmt.Printf("node status reported successfully\n")
}

// 更新CPU
func (this *NodeStatusExecutor) updateCPU(status *NodeStatus) {
	var duration = time.Duration(0)
	if this.isFirstTime {
		duration = 100 * time.Millisecond
	}

	percents, err := cpu.Percent(duration, false)
	if err != nil {
		status.Error = "cpu.Percent(): " + err.Error()
		return
	}
	if len(percents) == 0 {
		return
	}

	status.CPUUsage = percents[0] / 100

	if this.cpuLogicalCount == 0 && this.cpuPhysicalCount == 0 {
		status.CPULogicalCount, err = cpu.Counts(true)
		if err != nil {
			status.Error = "cpu.Counts(): " + err.Error()
			return
		}
		status.CPUPhysicalCount, err = cpu.Counts(false)
		if err != nil {
			status.Error = "cpu.Counts(): " + err.Error()
			return
		}
		this.cpuLogicalCount = status.CPULogicalCount
		this.cpuPhysicalCount = status.CPUPhysicalCount
	} else {
		status.CPULogicalCount = this.cpuLogicalCount
		status.CPUPhysicalCount = this.cpuPhysicalCount
	}
}
