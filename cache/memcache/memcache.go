package memcache

import (
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/model"
)

type Memcache struct {
}

func (c *Memcache) GetTorrents(key *common.TorrentParam, get model.TorrentObtainer) (torrents []model.Torrent, err error) {
	torrents, err = get()
	return
}

func (c *Memcache) ClearAll() {

}

func (c *Memcache) ClearTorrents() {

}

func New() *Memcache {
	return &Memcache{}
}
