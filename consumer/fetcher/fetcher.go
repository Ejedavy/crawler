package fetcher

type Fetcher interface {
	GetHTML(url string, headers map[string]string, proxy string) ([]byte, error)
}
