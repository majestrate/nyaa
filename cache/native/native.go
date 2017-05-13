package native

import (
	"container/list"
	"sync"
	"time"

	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/model"
)

const expiryTime = time.Minute

type torrentCache struct {
	data      map[common.TorrentParam]*list.Element
	l         *list.List
	totalUsed int
	mu        sync.Mutex
	parent    *NativeCache
}

// NativeCache implements cache.Cache
type NativeCache struct {
	torrents torrentCache
	// Size sets the maximum size of the cache before evicting unread data in MB
	Size float64
}

// New Creates New Native Cache instance
func New(sz float64) (c *NativeCache) {
	c = &NativeCache{
		torrents: torrentCache{
			data: make(map[common.TorrentParam]*list.Element, 10),
			l:    list.New(),
		},
		Size: sz,
	}
	c.torrents.parent = c
	return
}

// Key stores the ID of either a thread or board page
type Key struct {
	LastN uint8
	Board string
	ID    uint64
}

// Single cache entry for torrents
type torrentStore struct {
	sync.Mutex  // Controls general access to the contents of the struct
	lastFetched time.Time
	key         common.TorrentParam
	data        []model.Torrent
	count, size int
	parent      *torrentCache
}

// Check the cache for and existing record. If miss, run fn to retrieve fresh
// values.
func (n *NativeCache) GetTorrents(key *common.TorrentParam, get model.TorrentObtainer) (data []model.Torrent, err error) {
	s := n.getTorrentStore(key)

	// Also keeps multiple requesters from simultaneously requesting the same
	// data
	s.Lock()

	if s.isFresh() {
		data = s.data
	} else {
		data, err = get()
		if err == nil {
			s.update(data)
		} else {
			data = nil
		}
	}
	s.Unlock()
	return
}

// Retrieve a store from the cache or create a new one
func (n *NativeCache) getTorrentStore(key *common.TorrentParam) (s *torrentStore) {
	t := &n.torrents
	t.mu.Lock()

	k := key.Clone()
	el := t.data[k]
	if el == nil {
		s = &torrentStore{key: k, parent: t}
		t.data[k] = t.l.PushFront(s)
	} else {
		t.l.MoveToFront(el)
		s = el.Value.(*torrentStore)
	}

	return s
}

func (t *torrentCache) Clear() {
	t.mu.Lock()
	t.l = list.New()
	t.data = make(map[common.TorrentParam]*list.Element, 10)
	t.mu.Unlock()
}

// ClearTorrents clears all torrents from cache
func (n *NativeCache) ClearTorrents() {
	n.torrents.Clear()
}

// ClearAll clears the entire cache
func (n *NativeCache) ClearAll() {
	n.ClearTorrents()
}

// Update the total used memory counter and evict, if over limit
func (n *NativeCache) updateUsedSize(delta int) {
	t := &n.torrents
	t.mu.Lock()

	t.totalUsed += delta

	for t.totalUsed > int(n.Size)<<20 {
		e := t.l.Back()
		if e == nil {
			break
		}
		s := t.l.Remove(e).(*torrentStore)
		delete(t.data, s.key)
		t.totalUsed -= s.size
	}
	t.mu.Unlock()
}

// Return, if the data can still be considered fresh, without querying the DB
func (s *torrentStore) isFresh() bool {
	if s.lastFetched.IsZero() { // New store
		return false
	}
	return s.lastFetched.Add(expiryTime).After(time.Now())
}

// Stores the new values of s. Calculates and stores the new size. Passes the
// delta to the central cache to fire eviction checks.
func (s *torrentStore) update(data []model.Torrent) {
	newSize := 0
	for _, d := range data {
		newSize += d.Size()
	}
	s.data = data
	s.count = len(data)
	delta := newSize - s.size
	s.size = newSize
	s.lastFetched = time.Now()

	// Technically it is possible to update the size even when the store is
	// already evicted, but that should never happen, unless you have a very
	// small cache, very large stored datasets and a lot of traffic.
	s.parent.parent.updateUsedSize(delta)
}
