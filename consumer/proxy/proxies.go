package proxy

import "math/rand"

type ProxyGenFunc func(proxyType ProxyType) string
type ProxyType string

const (
	FreeProxy ProxyType = "free"
)

var proxyList = map[ProxyType][]string{
	FreeProxy: {
		"244.178.44.111:80",
		"89.0.142.86:80",
		"237.84.2.178:80",
		"89.207.132.170:80",
		"237.84.2.178:80",
		"38.0.101.76:80",
		"237.84.2.178:80",
		"244.178.44.111:80",
	},
}

func NewProxyGenFunc() ProxyGenFunc {
	return func(proxyType ProxyType) string {
		switch proxyType {
		case FreeProxy:
			idx := rand.Intn(len(proxyList[FreeProxy]))
			return proxyList[FreeProxy][idx]
		default:
			return ""
		}
	}
}
