# Network Probe (Go Version)

A comprehensive network testing tool written in Go that supports ping, tcping, traceroute, DNS queries, website testing, and port scanning.

## Features

- **ICMP Ping**: Test network connectivity using ICMP echo requests
- **TCP Ping**: Test TCP connection to specific ports
- **Website Testing**: Test website availability and response time
- **Traceroute**: Trace the path packets take to a destination
- **DNS Queries**: Perform various DNS record lookups
- **Port Scanning**: Scan for open ports on a target host
- **API Server**: Start a web API server for network testing

## Prerequisites

- Go 1.20 or later
- For ICMP ping and traceroute: root privileges (sudo)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/your-username/network-probe.git
cd network-probe
```

2. Build the project:

```bash
go build -o network_probe ./cmd/network-probe
```

## Usage

### Ping

```bash
sudo ./network_probe ping www.baidu.com
```

### TCP Ping

```bash
./network-probe tcping example.com -p 80
```

### Website Test

```bash
./network-probe website https://example.com
```

### Traceroute

```bash
sudo ./network-probe traceroute example.com
```

### DNS Query

```bash
./network-probe dns example.com -t A
```

### Port Scan

```bash
./network-probe port-scan example.com -r 1-1000
```

### Start API Server

```bash
./network-probe server --host 127.0.0.1 --port 8080
```

## Command Line Options

Run `./network-probe --help` to see all available commands and options.

## API Endpoints

When running the API server, the following endpoints are available:

- `POST /api/ping` - ICMP ping test
- `POST /api/tcping` - TCP connection test
- `POST /api/website` - Website test
- `POST /api/traceroute` - Traceroute
- `POST /api/dns` - DNS query
- `GET /api/health` - Health check
- `GET /api/status` - Service status
- `GET /ws` - WebSocket endpoint

## License

MIT
