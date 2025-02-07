package store

type Store interface {
	StoreContent(content []byte) error
}