package main

import (
	"fmt"
	"wishlist-bot/scrapers/amazonscraper"
)

type AmazonProductRepository interface {
	Save(id string, product *amazonscraper.Product) error
	Get(id string) (*amazonscraper.Product, error)
}

//Generic Memory storage
type MemoryStorage struct {
	storage map[string]interface{}
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		storage: make(map[string]interface{}),
	}
}

func (r *MemoryStorage) Save(id string, obj interface{}) error {
	r.storage[id] = obj
	return nil
}

func (r *MemoryStorage) Get(id string) (interface{}, error) {
	found := r.storage[id]

	var error error = nil

	if found == nil {
		error = fmt.Errorf("Not found")
	}
	return found, error
}

//Adapter
type MemoryStorageToRepositoryAdapter struct {
	MemoryStorage *MemoryStorage
}

func (r *MemoryStorageToRepositoryAdapter) Save(id string, product *amazonscraper.Product) error {
	return r.MemoryStorage.Save(id, *product)
}

func (r *MemoryStorageToRepositoryAdapter) Get(id string) (*amazonscraper.Product, error) {
	i, e := r.MemoryStorage.Get(id)
	if e != nil {
		return nil, e
	}
	p := i.(amazonscraper.Product)
	return &p, e
}
