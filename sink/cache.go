// Copyright © 2022 Sloan Childers
package sink

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

type FastCache struct {
	cache    *cache.Cache
	fileName string
}

func NewFastCache(fileName string) *FastCache {
	return &FastCache{
		cache:    cache.New(24*time.Hour, 60*time.Minute),
		fileName: fileName}
}

func (x *FastCache) Get(key string) (interface{}, bool) {
	return x.cache.Get(key)
}

func (x *FastCache) Set(key string, value interface{}, duration time.Duration) {
	x.cache.Set(key, value, duration)
}

func (x *FastCache) LoadFile() {
	x.cache.LoadFile(x.fileName)
}

func (x *FastCache) SaveFile() {
	x.cache.SaveFile(x.fileName)
}

func (x *FastCache) Clear(pattern string) {
	for k := range x.cache.Items() {
		if strings.Contains(k, pattern) {
			x.cache.Delete(k)
		}
	}
}
