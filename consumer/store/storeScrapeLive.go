package store

import (
	"consumer/models"
	"encoding/json"

	"github.com/go-redis/redis"
)

type ScrapeLiveStore struct {
	rdb *redis.Client
}

func NewScrapeLiveStore() Store {
	return &ScrapeLiveStore{
		rdb: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
	}
}

func (s *ScrapeLiveStore) StoreContent(content []byte) error {
	var product models.ScapeMeProduct
	json.Unmarshal(content, &product)
	res := s.rdb.HSet("scrapeLive:products", product.ID, content)
	return res.Err()
}
