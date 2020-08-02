package main

import (
	"fmt"
	"time"
	"wishlist-bot/chatapp"
)

//ProductRepo stores (Save) and retrieves (Get) chatapp.Product instances
type ProductRepo interface {
	Save(id string, p *chatapp.Product) error
	Get(id string) (*chatapp.Product, error)
}

type cacheItem struct {
	ts      time.Time
	product *chatapp.Product
}
type cacheRepo struct {
	duration time.Duration
	storage  map[string]*cacheItem
}

func newCacheRepo(itemDuration time.Duration) *cacheRepo {
	return &cacheRepo{
		duration: itemDuration,
		storage:  make(map[string]*cacheItem),
	}
}

func (r *cacheRepo) Save(id string, p *chatapp.Product) error {
	r.storage[id] = &cacheItem{
		product: p,
		ts:      time.Now(),
	}
	return nil
}

func (r *cacheRepo) Get(id string) (*chatapp.Product, error) {
	item := r.storage[id]
	if item == nil {
		return nil, fmt.Errorf("[CacheRepo] ID not found: %s", id)
	}
	if time.Now().Sub(item.ts) >= r.duration {
		return item.product, fmt.Errorf("[CacheRepo] ID Expired: %s", id)
	}
	return item.product, nil
}
