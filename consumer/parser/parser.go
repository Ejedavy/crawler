package parser

import (
	"consumer/fetcher"
	"consumer/header"
	"consumer/proxy"
)

type Parser interface {
	ExtractContent(content []byte) (interface{}, error)
	StoreContent(content []byte) error
	IsValidURL(url string) bool
	GetHTML(url string, headerGenFunc header.HeaderGenFunc, proxyGenFunc proxy.ProxyGenFunc, fetcher fetcher.Fetcher) ([]byte, error)
}
