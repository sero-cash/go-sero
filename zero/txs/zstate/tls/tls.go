package tls

import (
	"bytes"
	"github.com/sero-cash/go-czero-import/c_type"
	"runtime"
	"strconv"
	"sync"
)

var cache sync.Map

type PathsCache struct {
	Paths []c_type.Uint256
	Index    int
	Mining   bool
}

func Set(value *PathsCache) {
	id := GetGID()
	cache.Store(id, value)
}

func Get() *PathsCache {
	id := GetGID()
	if val, ok := cache.Load(id); ok {
		return val.(*PathsCache)
	} else {
		return &PathsCache{}
	}
}

func (cache *PathsCache) Add(value c_type.Uint256) {
	cache.Paths = append(cache.Paths, value)
}

func (cache *PathsCache) Next() (bool, c_type.Uint256) {
	defer func() {
		cache.Index += 1
	}()
	if len(cache.Paths) > cache.Index {
		return true, cache.Paths[cache.Index]
	}

	return false, c_type.Uint256{}
}

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
