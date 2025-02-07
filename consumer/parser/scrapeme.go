package parser

import (
	"consumer/fetcher"
	"consumer/header"
	"consumer/models"
	"consumer/proxy"
	"consumer/store"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ScrapemeParser struct {
	store store.Store
}

func NewScrapemeParser() Parser {
	return &ScrapemeParser{
		store: store.NewScrapeLiveStore(),
	}
}

func (s *ScrapemeParser) ExtractContent(content []byte) (interface{}, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		log.Printf("Failed to parse HTML: %v\n", err)
	}

	var products []models.ScapeMeProduct

	doc.Find(".product").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Find("a[data-product_id]").Attr("data-product_id")
		name := s.Find("h2").Text()
		price := s.Find(".amount").Text()

		products = append(products, models.ScapeMeProduct{
			ID:    id,
			Name:  strings.TrimSpace(name),
			Price: strings.TrimSpace(price),
		})
	})

	return products, nil
}

func (s *ScrapemeParser) IsValidURL(url string) bool {
	return true
}

func (s *ScrapemeParser) GetHTML(url string, headerGenFunc header.HeaderGenFunc, proxyGenFunc proxy.ProxyGenFunc, fetcher fetcher.Fetcher) ([]byte, error) {
	header := headerGenFunc()
	proxy := proxyGenFunc(proxy.FreeProxy)
	fmt.Println(header, proxy)
	return fetcher.GetHTML(url, header, proxy)
}

func (s *ScrapemeParser) StoreContent(content []byte) error {
	return s.store.StoreContent(content)
}
