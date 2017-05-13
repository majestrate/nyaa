package nop

import (
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/model"
)

type NopCache struct {
}

func (c *NopCache) GetTorrents(key *common.TorrentParam, get model.TorrentObtainer) ([]model.Torrent, error) {
	return get()
}

func (c *NopCache) ClearAll() {

}

func (c *NopCache) ClearTorrents() {

}

// New creates a new Cache that does NOTHING :D
func New() *NopCache {
	return &NopCache{}
}
