package cache

import (
	"github.com/ewhal/nyaa/cache/memcache"
	"github.com/ewhal/nyaa/cache/native"
	"github.com/ewhal/nyaa/cache/nop"
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/model"

	"errors"
)

// Cache defines interface for caching models
type Cache interface {
	GetTorrents(key *common.TorrentParam, get model.TorrentObtainer) ([]model.Torrent, error)
	ClearTorrents()
	ClearAll()
}

var ErrInvalidCacheDialect = errors.New("invalid cache dialect")

// Impl cache implementation instance
var Impl Cache

func Configure(conf *config.CacheConfig) (err error) {
	switch conf.Dialect {
	case "native":
		Impl = native.New(conf.Size)
		return
	case "memcache":
		Impl = memcache.New()
		return
	default:
		Impl = nop.New()
	}
	return
}
