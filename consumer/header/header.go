package header

import (
	"math/rand"
)

var headerList = []map[string]string{
	{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/237.84.2.178 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language": "en-GB,en;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
		"Connection":      "keep-alive",
	},
	{
		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "en-US,en;q=0.9",
		"Cache-Control":   "no-cache",
		"Connection":      "keep-alive",
	},
	{
		"User-Agent":                "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Accept-Encoding":           "gzip, deflate, br",
		"Accept-Language":           "en-US,en;q=0.5",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	},
	{
		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language": "en-US,en;q=0.9",
		"Cache-Control":   "max-age=0",
		"Connection":      "keep-alive",
		"Sec-Ch-Ua":       `"Chromium";v="92", " Not A;Brand";v="99", "Google Chrome";v="92"`,
	},
}

type HeaderGenFunc func() map[string]string

func NewHeaderGenFunc() HeaderGenFunc {
	return func() map[string]string {
		idx := rand.Intn(len(headerList))
		return headerList[idx]
	}
}
