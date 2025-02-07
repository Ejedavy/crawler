package fetcher

import "github.com/gocolly/colly"

type CollyFetcher struct {
	c *colly.Collector
}

func NewCollyFetcher() *CollyFetcher {
	return &CollyFetcher{
		c: colly.NewCollector(),
	}
}

func (c *CollyFetcher) GetHTML(url string, headers map[string]string, proxy string) ([]byte, error) {
	c.c.SetProxy(proxy)
	if userAgent, ok := headers["User-Agent"]; ok {
		c.c.UserAgent = userAgent
	}
	c.c.OnRequest(func(r *colly.Request) {
		for k, v := range headers {
			r.Headers.Set(k, v)
		}
	})
	var body []byte
	c.c.OnResponse(func(r *colly.Response) {
		body = r.Body
	})

	if err := c.c.Visit(url); err != nil {
		return nil, err
	}

	return body, nil
}
