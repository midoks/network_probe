package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"network-probe/internal/config"
	"network-probe/internal/modules"
	"network-probe/internal/utils/logger"
	"network-probe/internal/utils/report"
	"network-probe/internal/version"
)

// parseDuration 将秒数转换为 time.Duration
func parseDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}

// WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

var activeConnections int64

// WebSocketMessage 表示 WebSocket 消息
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	ID      string          `json:"id,omitempty"` // 消息 ID，用于跟踪请求
}

// WebSocketResponse 表示 WebSocket 响应
type WebSocketResponse struct {
	Type    string      `json:"type"`
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	ID      string      `json:"id,omitempty"` // 对应请求的消息 ID
}

// WebSocketUpdate 表示 WebSocket 实时更新
type WebSocketUpdate struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	ID   string      `json:"id,omitempty"` // 对应请求的消息 ID
}

// Server 表示 API 服务器
type Server struct {
	router *gin.Engine
	config *config.Config
}

// NewServer 创建一个新的 API 服务器
func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 加载配置
	cfg, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		fmt.Printf("Warning: failed to load config: %v, using defaults\n", err)
		cfg = &config.Config{
			Port:            8080,
			NodeID:          "default",
			Secret:          "",
			ReportEndpoints: []string{},
		}
	}

	// 配置 CORS
	router.Use(cors.Default())

	server := &Server{
		router: router,
		config: cfg,
	}

	// 设置路由
	server.setupRoutes()

	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	s.router.Use(s.connectionCounterMiddleware())

	// 健康检查（不需要认证）
	s.router.GET("/api/health", s.healthCheck)
	s.router.GET("/api/status", s.status)
	// 版本信息（不需要认证）
	s.router.GET("/api/version", s.version)

	// API 路由组（需要认证）
	api := s.router.Group("/api")
	api.Use(s.authMiddleware())
	{
		api.POST("/ping", s.handlePing)
		api.POST("/tcping", s.handleTcping)
		api.POST("/website", s.handleWebsite)
		api.POST("/traceroute", s.handleTraceroute)
		api.POST("/dns", s.handleDns)
		api.POST("/mtr", s.handleMtr)
	}

	s.router.GET("/api/stats", func(c *gin.Context) {
		current := atomic.LoadInt64(&activeConnections)
		c.JSON(http.StatusOK, gin.H{
			"active_connections": current,
		})
		// c.String(http.StatusOK, fmt.Sprintf("%d", current))
	})

	// 根路径
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Network Probe API Server")
	})

	s.router.GET("/ping", func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.String(http.StatusOK, "pong")
	})

	// WebSocket 路由（需要认证）
	s.router.GET("/ws", s.handleWebSocket)
}

// 连接统计中间件
func (s *Server) connectionCounterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求到来时计数器+1
		atomic.AddInt64(&activeConnections, 1)
		// 确保在请求结束时计数器-1
		defer atomic.AddInt64(&activeConnections, -1)
		c.Next()
	}
}

// authMiddleware 认证中间件
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果没有配置 secret，则跳过认证
		if s.config.Secret == "" {
			c.Next()
			return
		}

		// 从请求头获取认证信息
		nodeID := c.GetHeader("X-Node-ID")
		secret := c.GetHeader("X-Secret")

		// 验证认证信息
		if nodeID != s.config.NodeID || secret != s.config.Secret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// status 服务状态
func (s *Server) status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "running",
		"version": version.Version,
	})
}

// version 版本信息
func (s *Server) version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":     version.Version,
		"build_time":  version.BuildTime,
		"service":     "Network Probe",
		"api_version": "v1",
	})
}

// handlePing 处理 ping 请求
func (s *Server) handlePing(c *gin.Context) {
	var req struct {
		Host    string `json:"host" binding:"required"`
		Count   int    `json:"count"`
		Timeout int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 4
	}
	if req.Timeout == 0 {
		req.Timeout = 2
	}

	service := modules.NewPingService()
	config := modules.NewPingConfig()
	config.Host = req.Host
	config.Count = req.Count
	config.Timeout = parseDuration(req.Timeout)

	result, err := service.Ping(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "ping",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report ping: %v\n", err)
	}

	c.JSON(http.StatusOK, result)
}

// handleTcping 处理 tcping 请求
func (s *Server) handleTcping(c *gin.Context) {
	var req struct {
		Host    string `json:"host" binding:"required"`
		Port    int    `json:"port" binding:"required"`
		Count   int    `json:"count"`
		Timeout int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 4
	}
	if req.Timeout == 0 {
		req.Timeout = 3
	}

	service := modules.NewTcpingService()
	config := modules.NewTcpingConfig()
	config.Host = req.Host
	config.Port = req.Port
	config.Count = req.Count
	config.Timeout = parseDuration(req.Timeout)

	result, err := service.Tcping(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "tcping",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report tcping: %v\n", err)
	}

	c.JSON(http.StatusOK, result)
}

// handleWebsite 处理 website 请求
func (s *Server) handleWebsite(c *gin.Context) {
	var req struct {
		URL             string `json:"url" binding:"required"`
		Method          string `json:"method"`
		Timeout         int    `json:"timeout"`
		FollowRedirects bool   `json:"follow_redirects"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	service := modules.NewWebsiteTestService()
	config := modules.NewWebsiteTestConfig()
	config.URL = req.URL
	config.Method = req.Method
	config.Timeout = parseDuration(req.Timeout)
	config.FollowRedirects = req.FollowRedirects

	result, err := service.TestWebsite(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "website",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report website: %v\n", err)
	}

	c.JSON(http.StatusOK, result)
}

// handleTraceroute 处理 traceroute 请求
func (s *Server) handleTraceroute(c *gin.Context) {
	var req struct {
		Host     string `json:"host" binding:"required"`
		MaxHops  int    `json:"max_hops"`
		Protocol string `json:"protocol"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.MaxHops == 0 {
		req.MaxHops = 30
	}
	if req.Protocol == "" {
		req.Protocol = "icmp"
	}

	service := modules.NewTracerouteService()
	config := modules.NewTracerouteConfig()
	config.Host = req.Host
	config.MaxHops = req.MaxHops
	config.Protocol = req.Protocol

	result, err := service.Traceroute(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "traceroute",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report traceroute: %v\n", err)
	}

	c.JSON(http.StatusOK, result)
}

// handleDns 处理 DNS 请求
func (s *Server) handleDns(c *gin.Context) {
	var req struct {
		Domain     string `json:"domain" binding:"required"`
		QueryType  string `json:"query_type"`
		Nameserver string `json:"nameserver"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.QueryType == "" {
		req.QueryType = "A"
	}

	var service *modules.DnsService
	var err error

	if req.Nameserver != "" {
		service, err = modules.NewDnsServiceWithNameserver(req.Nameserver)
	} else {
		service = modules.NewDnsService()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config := modules.NewDnsConfig()
	config.Domain = req.Domain
	config.QueryType = modules.DnsQueryType(req.QueryType)
	config.Nameserver = req.Nameserver

	result, err := service.Query(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "dns",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report dns: %v\n", err)
	}

	c.JSON(http.StatusOK, result)
}

// handleMtr 处理 mtr 请求
func (s *Server) handleMtr(c *gin.Context) {
	var req struct {
		Host     string `json:"host" binding:"required"`
		MaxHops  int    `json:"max_hops"`
		Count    int    `json:"count"`
		Interval int    `json:"interval"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.MaxHops == 0 {
		req.MaxHops = 30
	}
	if req.Count == 0 {
		req.Count = 10
	}
	if req.Interval == 0 {
		req.Interval = 1
	}

	service := modules.NewMtrService()
	config := modules.NewMtrConfig()
	config.Host = req.Host
	config.MaxHops = req.MaxHops
	config.Count = req.Count
	config.Interval = req.Interval

	result, err := service.Mtr(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "mtr",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report mtr: %v\n", err)
	}

	c.JSON(http.StatusOK, result)
}

// handleWebSocket 处理 WebSocket 连接
func (s *Server) handleWebSocket(c *gin.Context) {
	// 检查认证信息
	fmt.Println("s.config.Debug:", s.config.Debug)
	if s.config.Debug {
		if s.config.Secret != "" {
			nodeID := c.GetHeader("X-Node-ID")
			secret := c.GetHeader("X-Secret")

			if nodeID != s.config.NodeID || secret != s.config.Secret {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
		}
	}

	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket: " + err.Error()})
		return
	}
	defer conn.Close()

	// 增加活跃连接计数
	atomic.AddInt64(&activeConnections, 1)
	defer atomic.AddInt64(&activeConnections, -1)

	// 发送连接成功消息
	welcome := WebSocketResponse{
		Type:    "connected",
		Status:  "success",
		Message: "Connected to Network Probe WebSocket",
	}
	if err := conn.WriteJSON(welcome); err != nil {
		return
	}

	// 上报 WebSocket 连接
	if err := report.ReportRequest(map[string]interface{}{
		"tag":    "websocket_connect",
		"status": "connected",
	}); err != nil {
		fmt.Printf("Failed to report WebSocket connect: %v\n", err)
	}
	// 处理消息
	for {
		var msg WebSocketMessage
		if err := conn.ReadJSON(&msg); err != nil {
			// 客户端断开连接
			// 上报 WebSocket 断开连接
			if err := report.ReportRequest(map[string]interface{}{
				"tag":    "websocket_disconnect",
				"status": "disconnected",
			}); err != nil {
				fmt.Printf("Failed to report WebSocket disconnect: %v\n", err)
			}
			break
		}

		// 根据消息类型处理
		if msg.Type == "mtr" {
			// MTR 特殊处理，支持实时更新
			s.handleWebSocketMtrWithUpdates(conn, msg)
		} else {
			// 其他消息类型
			response := s.processWebSocketMessage(msg)
			response.ID = msg.ID // 保留消息 ID
			if err := conn.WriteJSON(response); err != nil {
				break
			}
		}
	}
}

// processWebSocketMessage 处理 WebSocket 消息
func (s *Server) processWebSocketMessage(msg WebSocketMessage) WebSocketResponse {
	switch msg.Type {
	case "ping":
		return s.handleWebSocketPing(msg.Payload)
	case "tcping":
		return s.handleWebSocketTcping(msg.Payload)
	case "website":
		return s.handleWebSocketWebsite(msg.Payload)
	case "traceroute":
		return s.handleWebSocketTraceroute(msg.Payload)
	case "dns":
		return s.handleWebSocketDns(msg.Payload)
	case "mtr":
		return s.handleWebSocketMtr(msg.Payload)
	default:
		return WebSocketResponse{
			Type:    msg.Type,
			Status:  "error",
			Message: "Unknown message type: " + msg.Type,
		}
	}
}

// handleWebSocketPing 处理 WebSocket ping 请求
func (s *Server) handleWebSocketPing(payload json.RawMessage) WebSocketResponse {
	var req struct {
		Host    string `json:"host"`
		Count   int    `json:"count"`
		Timeout int    `json:"timeout"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return WebSocketResponse{
			Type:    "ping",
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		}
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 4
	}
	if req.Timeout == 0 {
		req.Timeout = 2
	}

	service := modules.NewPingService()
	config := modules.NewPingConfig()
	config.Host = req.Host
	config.Count = req.Count
	config.Timeout = parseDuration(req.Timeout)

	result, err := service.Ping(config)
	if err != nil {
		return WebSocketResponse{
			Type:    "ping",
			Status:  "error",
			Message: err.Error(),
		}
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "websocket_ping",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report WebSocket ping: %v\n", err)
	}

	return WebSocketResponse{
		Type:   "ping",
		Status: "success",
		Data:   result,
	}
}

// handleWebSocketTcping 处理 WebSocket tcping 请求
func (s *Server) handleWebSocketTcping(payload json.RawMessage) WebSocketResponse {
	var req struct {
		Host    string `json:"host"`
		Port    int    `json:"port"`
		Count   int    `json:"count"`
		Timeout int    `json:"timeout"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return WebSocketResponse{
			Type:    "tcping",
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		}
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 4
	}
	if req.Timeout == 0 {
		req.Timeout = 3
	}

	service := modules.NewTcpingService()
	config := modules.NewTcpingConfig()
	config.Host = req.Host
	config.Port = req.Port
	config.Count = req.Count
	config.Timeout = parseDuration(req.Timeout)

	result, err := service.Tcping(config)
	if err != nil {
		return WebSocketResponse{
			Type:    "tcping",
			Status:  "error",
			Message: err.Error(),
		}
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "websocket_tcping",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report WebSocket tcping: %v\n", err)
	}

	return WebSocketResponse{
		Type:   "tcping",
		Status: "success",
		Data:   result,
	}
}

// handleWebSocketWebsite 处理 WebSocket website 请求
func (s *Server) handleWebSocketWebsite(payload json.RawMessage) WebSocketResponse {
	var req struct {
		URL             string `json:"url"`
		Method          string `json:"method"`
		Timeout         int    `json:"timeout"`
		FollowRedirects bool   `json:"follow_redirects"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return WebSocketResponse{
			Type:    "website",
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		}
	}

	// 设置默认值
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	service := modules.NewWebsiteTestService()
	config := modules.NewWebsiteTestConfig()
	config.URL = req.URL
	config.Method = req.Method
	config.Timeout = parseDuration(req.Timeout)
	config.FollowRedirects = req.FollowRedirects

	result, err := service.TestWebsite(config)
	if err != nil {
		return WebSocketResponse{
			Type:    "website",
			Status:  "error",
			Message: err.Error(),
		}
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "websocket_website",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report WebSocket website: %v\n", err)
	}

	return WebSocketResponse{
		Type:   "website",
		Status: "success",
		Data:   result,
	}
}

// handleWebSocketTraceroute 处理 WebSocket traceroute 请求
func (s *Server) handleWebSocketTraceroute(payload json.RawMessage) WebSocketResponse {
	var req struct {
		Host     string `json:"host"`
		MaxHops  int    `json:"max_hops"`
		Protocol string `json:"protocol"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return WebSocketResponse{
			Type:    "traceroute",
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		}
	}

	// 设置默认值
	if req.MaxHops == 0 {
		req.MaxHops = 30
	}
	if req.Protocol == "" {
		req.Protocol = "icmp"
	}

	service := modules.NewTracerouteService()
	config := modules.NewTracerouteConfig()
	config.Host = req.Host
	config.MaxHops = req.MaxHops
	config.Protocol = req.Protocol

	result, err := service.Traceroute(config)
	if err != nil {
		return WebSocketResponse{
			Type:    "traceroute",
			Status:  "error",
			Message: err.Error(),
		}
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "websocket_traceroute",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report WebSocket traceroute: %v\n", err)
	}

	return WebSocketResponse{
		Type:   "traceroute",
		Status: "success",
		Data:   result,
	}
}

// handleWebSocketDns 处理 WebSocket dns 请求
func (s *Server) handleWebSocketDns(payload json.RawMessage) WebSocketResponse {
	var req struct {
		Domain     string `json:"domain"`
		QueryType  string `json:"query_type"`
		Nameserver string `json:"nameserver"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return WebSocketResponse{
			Type:    "dns",
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		}
	}

	// 设置默认值
	if req.QueryType == "" {
		req.QueryType = "A"
	}

	var service *modules.DnsService
	var err error

	if req.Nameserver != "" {
		service, err = modules.NewDnsServiceWithNameserver(req.Nameserver)
	} else {
		service = modules.NewDnsService()
	}

	if err != nil {
		return WebSocketResponse{
			Type:    "dns",
			Status:  "error",
			Message: err.Error(),
		}
	}

	config := modules.NewDnsConfig()
	config.Domain = req.Domain
	config.QueryType = modules.DnsQueryType(req.QueryType)
	config.Nameserver = req.Nameserver

	result, err := service.Query(config)
	if err != nil {
		return WebSocketResponse{
			Type:    "dns",
			Status:  "error",
			Message: err.Error(),
		}
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "websocket_dns",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report WebSocket dns: %v\n", err)
	}

	return WebSocketResponse{
		Type:   "dns",
		Status: "success",
		Data:   result,
	}
}

// handleWebSocketMtr 处理 WebSocket mtr 请求
func (s *Server) handleWebSocketMtr(payload json.RawMessage) WebSocketResponse {
	var req struct {
		Host     string `json:"host"`
		MaxHops  int    `json:"max_hops"`
		Count    int    `json:"count"`
		Interval int    `json:"interval"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		return WebSocketResponse{
			Type:    "mtr",
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		}
	}

	// 设置默认值
	if req.MaxHops == 0 {
		req.MaxHops = 30
	}
	if req.Count == 0 {
		req.Count = 10
	}
	if req.Interval == 0 {
		req.Interval = 1
	}

	service := modules.NewMtrService()
	config := modules.NewMtrConfig()
	config.Host = req.Host
	config.MaxHops = req.MaxHops
	config.Count = req.Count
	config.Interval = req.Interval

	result, err := service.Mtr(config)
	if err != nil {
		return WebSocketResponse{
			Type:    "mtr",
			Status:  "error",
			Message: err.Error(),
		}
	}

	// 上报结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "websocket_mtr",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report WebSocket mtr: %v\n", err)
	}

	return WebSocketResponse{
		Type:   "mtr",
		Status: "success",
		Data:   result,
	}
}

// handleWebSocketMtrWithUpdates 处理 WebSocket mtr 请求（带实时更新）
func (s *Server) handleWebSocketMtrWithUpdates(conn *websocket.Conn, msg WebSocketMessage) {
	var req struct {
		Host     string `json:"host"`
		MaxHops  int    `json:"max_hops"`
		Count    int    `json:"count"`
		Interval int    `json:"interval"`
	}

	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		response := WebSocketResponse{
			Type:    "mtr",
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
			ID:      msg.ID,
		}
		conn.WriteJSON(response)
		return
	}

	// 设置默认值
	if req.MaxHops == 0 {
		req.MaxHops = 30
	}
	if req.Count == 0 {
		req.Count = 10
	}
	if req.Interval == 0 {
		req.Interval = 1
	}

	// 上报 MTR 请求开始
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "websocket_mtr_start",
		"request": req,
	}); err != nil {
		fmt.Printf("Failed to report WebSocket mtr start: %v\n", err)
	}
	// 创建带有回调函数的 MTR 服务
	hopCallback := func(hop modules.MtrHop) error {
		// 发送跳点更新
		update := WebSocketUpdate{
			Type: "mtr_update",
			Data: hop,
			ID:   msg.ID,
		}
		return conn.WriteJSON(update)
	}

	// 数据包回调函数
	packetCallback := func(packet modules.MtrPacketResult) error {
		// 发送数据包更新
		update := WebSocketUpdate{
			Type: "mtr_packet_update",
			Data: packet,
			ID:   msg.ID,
		}
		return conn.WriteJSON(update)
	}

	service := modules.NewMtrServiceWithPacketCallback(hopCallback, packetCallback)
	config := modules.NewMtrConfig()
	config.Host = req.Host
	config.MaxHops = req.MaxHops
	config.Count = req.Count
	config.Interval = req.Interval

	// 执行 MTR
	result, err := service.Mtr(config)
	if err != nil {
		response := WebSocketResponse{
			Type:    "mtr",
			Status:  "error",
			Message: err.Error(),
			ID:      msg.ID,
		}
		conn.WriteJSON(response)
		return
	}

	// 上报 MTR 结果
	if err := report.ReportRequest(map[string]interface{}{
		"tag":     "websocket_mtr_complete",
		"request": req,
		"result":  result,
	}); err != nil {
		fmt.Printf("Failed to report WebSocket mtr complete: %v\n", err)
	}

	// 发送最终结果
	response := WebSocketResponse{
		Type:   "mtr",
		Status: "success",
		Data:   result,
		ID:     msg.ID,
	}
	conn.WriteJSON(response)
}

// GetConfig 获取服务器配置
func (s *Server) GetConfig() *config.Config {
	return s.config
}

// Run 启动服务器
func (s *Server) Run(addr string) error {
	logger.RewriteStderrFile()
	//test

	// api server
	go func() {
		if err := s.router.Run(addr); err != nil {
			fmt.Printf("failed bind port error: %v\n", err)
			// panic("bind panic error:" + fmt.Sprintf("panic error:%s", addr))
		}
	}()

	// report
	if err := report.NodeInfo("node", "starting ..."); err != nil {
		fmt.Printf("Failed to report node startup: %v\n", err)
	}

	// 启动节点状态上报
	go func() {
		report.NewNodeStatusExecutor().Run()
	}()

	select {}
}
