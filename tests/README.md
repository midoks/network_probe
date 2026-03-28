# WebSocket 测试客户端

这是一个用于测试 Network Probe WebSocket 功能的 Go 客户端。

## 功能特性

- 支持 WebSocket 连接和认证
- 测试所有网络探测功能：MTR、Ping、TCPing、Website、DNS
- 详细的请求和响应日志
- 错误处理和状态检查
- 支持环境变量配置认证信息

## 使用方法

### 基本使用

```bash
go run tests/websocket_client.go
```

### 使用自定义认证信息

```bash
NODE_ID="your-node-id" SECRET="your-secret" go run tests/websocket_client.go
```

### 环境变量

- `NODE_ID`: 节点 ID（默认：xxx）
- `SECRET`: 密钥（默认：xxx）

## 测试内容

客户端会依次测试以下功能：

1. **MTR 测试**: 测试到 baidu.com 的网络路径
2. **Ping 测试**: ICMP ping 测试
3. **TCPing 测试**: TCP 连接测试
4. **Website 测试**: 网站 HTTP 测试
5. **DNS 测试**: DNS 查询测试

## 测试结果示例

```
Connected to WebSocket
Using Node ID: xxx
Server: {Type:connected Status:success Message:Connected to Network Probe WebSocket Data:<nil>}

=== Testing MTR ===
Sending:
{
  "payload": {
    "count": 3,
    "host": "baidu.com",
    "interval": 1,
    "max_hops": 5
  },
  "type": "mtr"
}
Response:
{
  "type": "mtr",
  "status": "error",
  "message": "mtr requires root privileges. Please run with sudo or as root"
}
⚠️  Error: mtr requires root privileges. Please run with sudo or as root

=== Testing TCPing ===
Sending:
{
  "payload": {
    "count": 2,
    "host": "baidu.com",
    "port": 80,
    "timeout": 3
  },
  "type": "tcping"
}
Response:
{
  "type": "tcping",
  "status": "success",
  "data": {
    "Attempts": 2,
    "AvgRTT": 151.6211455,
    "Host": "baidu.com",
    "IP": "124.237.177.164",
    "MaxRTT": 176.22574999999998,
    "MinRTT": 127.016541,
    "PacketLoss": 0,
    "Port": 80,
    "SuccessfulAttempts": 2
  }
}
✅ Success

=== All tests completed ===
```

## 依赖

- Go 1.20+
- github.com/gorilla/websocket

## 注意事项

1. 确保 Network Probe 服务器正在运行
2. 默认连接到 `ws://127.0.0.1:8081/ws`
3. 某些功能（如 Ping 和 MTR）需要 root 权限
4. 测试超时设置为 30 秒
