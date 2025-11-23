package utils

import (
	"errors"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// GetAvailablePort 会尝试获取可用的监听地址
func GetAvailablePort() (string, error) {
	if config.Service.AddrList == nil {
		return "", errors.New("utils.GetAvailablePort: config.Service.AddrList is nil")
	}
	for _, addr := range config.Service.AddrList {
		if ok := AddrCheck(addr); ok {
			if !strings.HasPrefix(addr, "0.0.0.0") {
				return addr, nil
			}
			return getOutboundIP() + ":" + strconv.Itoa(AddrGetPort(addr)), nil
		}
	}
	return "", errors.New("utils.GetAvailablePort: not available port from config")
}

// AddrCheck 会检查当前的监听地址是否已被占用
func AddrCheck(addr string) bool {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	defer func() {
		if err := l.Close(); err != nil {
			logger.Errorf("utils.AddrCheck: failed to close listener: %v", err.Error())
		}
	}()
	return true
}
func AddrGetPort(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return 0
	}
	return port
}

func AddrGetHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return host
}

// getOutboundIP 返回当前主机的出网 IP（即系统默认的外部通信地址）
func getOutboundIP() string {
	// 创建一个 UDP 连接，目标是任意外部地址（这里使用 8.8.8.8:80）。
	// 连接不会真正建立，只是触发系统选择出网网卡。
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// 如果无法建立连接（例如无网络），返回 127.0.0.1 作为兜底地址。
		return "127.0.0.1"
	}
	defer conn.Close()

	// 获取本地连接地址（LocalAddr），其中包含了系统为该连接选择的 IP 和端口。
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// 返回本机出网 IP。
	return localAddr.IP.String()
}

// ParseCookiesToString 会尝试解析 cookies 到 string
// 只会返回 "Key=Value; Key=Value"样式的文本数组
func ParseCookiesToString(cookies []*http.Cookie) string {
	var cookieStrings []string
	for _, cookie := range cookies {
		cookieStrings = append(cookieStrings, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(cookieStrings, "; ")
}

// ParseCookies 将cookie字符串解析为 http.Cookie
// 这里只能解析这样的数组： "Key=Value; Key=Value"
func ParseCookies(rawData string) []*http.Cookie {
	var cookies []*http.Cookie
	maxSplitNumber := 2

	// 按照分号分割每个 Cookie
	pairs := strings.Split(rawData, ";")
	for _, pair := range pairs {
		// 去除空格并检查是否为空
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// 按等号分割键和值
		parts := strings.SplitN(pair, "=", maxSplitNumber)
		if len(parts) != maxSplitNumber {
			continue // 如果格式不正确，跳过
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 创建 http.Cookie 并添加到切片
		cookie := &http.Cookie{
			Name:  key,
			Value: value,
		}
		cookies = append(cookies, cookie)
	}

	return cookies
}
