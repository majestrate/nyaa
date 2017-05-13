package scraperService

import (
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/util/log"
	"net"
	"net/url"
	"time"
)

// MTU yes this is the ipv6 mtu
const MTU = 1500

// max number of scrapes per packet
const ScrapesPerPacket = 74

// bittorrent scraper
type Scraper struct {
	done             chan int
	sendQueue        chan *SendEvent
	recvQueue        chan *RecvEvent
	errQueue         chan error
	trackers         map[string]*Bucket
	ticker           *time.Ticker
	cleanup          *time.Ticker
	interval         time.Duration
	PacketsPerSecond uint
}

func New(conf *config.ScraperConfig) (sc *Scraper, err error) {
	sc = &Scraper{
		done:      make(chan int),
		sendQueue: make(chan *SendEvent, 1024),
		recvQueue: make(chan *RecvEvent, 1024),
		errQueue:  make(chan error),
		trackers:  make(map[string]*Bucket),
		ticker:    time.NewTicker(time.Second * 10),
		interval:  time.Second * time.Duration(conf.IntervalSeconds),
		cleanup:   time.NewTicker(time.Minute),
	}

	if sc.PacketsPerSecond == 0 {
		sc.PacketsPerSecond = 10
	}

	for idx := range conf.Trackers {
		err = sc.AddTracker(&conf.Trackers[idx])
		if err != nil {
			break
		}
	}
	return
}

func (sc *Scraper) AddTracker(conf *config.ScrapeConfig) (err error) {
	var u *url.URL
	u, err = url.Parse(conf.URL)
	if err == nil {
		var ips []net.IP
		ips, err = net.LookupIP(u.Hostname())
		if err == nil {
			// TODO: use more than 1 ip ?
			addr := &net.UDPAddr{
				IP: ips[0],
			}
			addr.Port, err = net.LookupPort("udp", u.Port())
			if err == nil {
				sc.trackers[addr.String()] = NewBucket(addr)
			}
		}
	}
	return
}

func (sc *Scraper) Close() (err error) {
	close(sc.sendQueue)
	close(sc.recvQueue)
	close(sc.errQueue)
	sc.ticker.Stop()
	sc.done <- 1
	return
}

func (sc *Scraper) runRecv(pc net.PacketConn) {
	for {
		var buff [MTU]byte
		n, from, err := pc.ReadFrom(buff[:])

		if err == nil {

			log.Debugf("got %d from %s", n, from)
			sc.recvQueue <- &RecvEvent{
				From: from,
				Data: buff[:n],
			}
		} else {
			sc.errQueue <- err
		}
	}
}

func (sc *Scraper) runSend(pc net.PacketConn) {
	for {
		ev, ok := <-sc.sendQueue
		if !ok {
			return
		}
		log.Debugf("write %d to %s", len(ev.Data), ev.To)
		pc.WriteTo(ev.Data, ev.To)
	}
}

func (sc *Scraper) RunWorker(pc net.PacketConn) (err error) {

	go sc.runRecv(pc)
	go sc.runSend(pc)
	for {
		var bucket *Bucket
		ev, ok := <-sc.recvQueue
		if !ok {
			break
		}
		tid, err := ev.TID()
		action, err := ev.Action()
		log.Debugf("transaction = %d action = %d", tid, action)
		if err == nil {
			bucket, ok = sc.trackers[ev.From.String()]
			if ok && bucket != nil {
				bucket.VisitTransaction(tid, func(t *Transaction) {
					if t == nil {
						log.Warnf("no transaction %d", tid)
					} else {
						if t.GotData(ev.Data) {
							err := t.Sync()
							if err != nil {
								log.Warnf("failed to sync swarm: %s", err)
							}
							t.Done()
							log.Debugf("transaction %d done", tid)
						} else {
							sc.sendQueue <- t.SendEvent(ev.From)
						}
					}
				})
			} else {
				log.Warnf("bucket not found for %s", ev.From)
			}
		}

	}
	return
}

func (sc *Scraper) Run() {
	for {
		select {
		case <-sc.ticker.C:
			sc.Scrape(sc.PacketsPerSecond)
			break
		case <-sc.cleanup.C:
			sc.removeStale()
			break
		}
	}
}

func (sc *Scraper) removeStale() {

	for k := range sc.trackers {
		sc.trackers[k].ForEachTransaction(func(tid uint32, t *Transaction) {
			if t == nil || t.IsTimedOut() {
				sc.trackers[k].Forget(tid)
			}
		})
	}
}

func (sc *Scraper) Scrape(packets uint) {
	now := time.Now().Add(0 - sc.interval)
	// only scrape torretns uploaded within 90 days
	oldest := now.Add(0 - (time.Hour * 24 * 90))
	rows, err := db.Impl.Query("SELECT torrent_id, torrent_hash FROM torrents WHERE ( last_scrape IS NULL OR  last_scrape < $1 ) AND date > $2 ORDER BY torrent_id DESC LIMIT $3", now, oldest, packets*ScrapesPerPacket)
	if err == nil {
		counter := 0
		var scrape [ScrapesPerPacket]common.ScrapeResult
		for rows.Next() {
			idx := counter % ScrapesPerPacket
			rows.Scan(&scrape[idx].TorrentID, &scrape[idx].Hash)
			counter++
			if counter%ScrapesPerPacket == 0 {
				for _, b := range sc.trackers {
					t := b.NewTransaction(scrape[:])
					sc.sendQueue <- t.SendEvent(b.Addr)
				}
			}
		}
		idx := counter % ScrapesPerPacket
		if idx > 0 {
			for _, b := range sc.trackers {
				t := b.NewTransaction(scrape[:idx])
				sc.sendQueue <- t.SendEvent(b.Addr)
			}
		}
		log.Infof("scrape %d", counter)
		rows.Close()

	} else {
		log.Warnf("failed to select torrents for scrape: %s", err)
	}
}

func (sc *Scraper) Wait() {
	<-sc.done
}
