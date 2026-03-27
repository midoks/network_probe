package modules

import (
	"fmt"
	"net"
	"strings"
)

// DnsQueryType 表示 DNS 查询类型
type DnsQueryType string

// 常见的 DNS 查询类型
const (
	DnsTypeA     DnsQueryType = "A"
	DnsTypeAAAA  DnsQueryType = "AAAA"
	DnsTypeCNAME DnsQueryType = "CNAME"
	DnsTypeMX    DnsQueryType = "MX"
	DnsTypeNS    DnsQueryType = "NS"
	DnsTypeTXT   DnsQueryType = "TXT"
	DnsTypeSOA   DnsQueryType = "SOA"
)

// DnsConfig 表示 DNS 查询配置
type DnsConfig struct {
	Domain     string
	QueryType  DnsQueryType
	Nameserver string
}

// NewDnsConfig 创建一个新的 DNS 查询配置
func NewDnsConfig() *DnsConfig {
	return &DnsConfig{
		QueryType: DnsTypeA,
	}
}

// DnsRecord 表示 DNS 记录
type DnsRecord struct {
	Type  string
	Value string
	TTL   int
}

// DnsResult 表示 DNS 查询结果
type DnsResult struct {
	Domain  string
	Records []DnsRecord
	Error   string
}

// DnsService 表示 DNS 服务
type DnsService struct {
	nameserver string
}

// NewDnsService 创建一个新的 DNS 服务
func NewDnsService() *DnsService {
	return &DnsService{}
}

// NewDnsServiceWithNameserver 创建一个使用指定 nameserver 的 DNS 服务
func NewDnsServiceWithNameserver(nameserver string) (*DnsService, error) {
	// 验证 nameserver 格式
	if !strings.Contains(nameserver, ":") {
		nameserver = nameserver + ":53"
	}

	// 测试 nameserver 是否可达
	_, err := net.Dial("udp", nameserver)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nameserver: %v", err)
	}

	return &DnsService{
		nameserver: nameserver,
	}, nil
}

// Query 执行 DNS 查询
func (s *DnsService) Query(config *DnsConfig) (*DnsResult, error) {
	result := &DnsResult{
		Domain:  config.Domain,
		Records: make([]DnsRecord, 0),
	}

	// 根据查询类型执行不同的查询
	switch config.QueryType {
	case DnsTypeA:
		ips, err := net.LookupIP(config.Domain)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}
		for _, ip := range ips {
			if ip.To4() != nil {
				result.Records = append(result.Records, DnsRecord{
					Type:  "A",
					Value: ip.String(),
				})
			}
		}

	case DnsTypeAAAA:
		ips, err := net.LookupIP(config.Domain)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}
		for _, ip := range ips {
			if ip.To4() == nil {
				result.Records = append(result.Records, DnsRecord{
					Type:  "AAAA",
					Value: ip.String(),
				})
			}
		}

	case DnsTypeCNAME:
		cnames, err := net.LookupCNAME(config.Domain)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}
		result.Records = append(result.Records, DnsRecord{
			Type:  "CNAME",
			Value: cnames,
		})

	case DnsTypeMX:
		mx, err := net.LookupMX(config.Domain)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}
		for _, m := range mx {
			result.Records = append(result.Records, DnsRecord{
				Type:  "MX",
				Value: fmt.Sprintf("%s (priority: %d)", m.Host, m.Pref),
			})
		}

	case DnsTypeNS:
		ns, err := net.LookupNS(config.Domain)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}
		for _, n := range ns {
			result.Records = append(result.Records, DnsRecord{
				Type:  "NS",
				Value: n.Host,
			})
		}

	case DnsTypeTXT:
		txt, err := net.LookupTXT(config.Domain)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}
		for _, t := range txt {
			result.Records = append(result.Records, DnsRecord{
				Type:  "TXT",
				Value: t,
			})
		}

	default:
		result.Error = fmt.Sprintf("unsupported query type: %s", config.QueryType)
		return result, nil
	}

	return result, nil
}
