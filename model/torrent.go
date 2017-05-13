package model

import (
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/util"

	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"
)

type Feed struct {
	ID        uint32
	Name      string
	Hash      string
	Magnet    string
	Timestamp string
}

type Torrent struct {
	ID          uint32
	Name        string
	Hash        string
	Category    int
	SubCategory int
	Status      int
	Date        time.Time
	UploaderID  uint32
	Downloads   uint64
	Stardom     int
	Filesize    uint64
	Description string
	WebsiteLink string
	DeletedAt   *time.Time

	Seeders    uint32
	Leechers   uint32
	Completed  uint32
	LastScrape time.Time

	Comments []Comment
	Uploader *User
}

// Returns the total size of memory recursively allocated for this struct
// FIXME: doesn't go have sizeof or something nicer for this?
func (t Torrent) Size() (s int) {
	s += 8 + // ints
		2*3 + // time.Time
		2 + // pointers
		4*2 + // string pointers
		// string array sizes
		len(t.Name) + len(t.Hash) + len(t.Description) + len(t.WebsiteLink) +
		2*2 // array pointers
	s *= 8 // Assume 64 bit OS

	return

}

func (t *Torrent) MarshalJSON() ([]byte, error) {
	torrentLink := ""
	if t.ID <= config.LastOldTorrentID && len(config.TorrentCacheLink) > 0 {
		torrentLink = fmt.Sprintf(config.TorrentCacheLink, t.Hash)
	} else if t.ID > config.LastOldTorrentID && len(config.TorrentStorageLink) > 0 {
		torrentLink = fmt.Sprintf(config.TorrentStorageLink, t.Hash)
	}
	pubForm := map[string]interface{}{
		"id":            t.ID,
		"name":          t.Name,
		"status":        t.Status,
		"hash":          t.Hash,
		"date":          t.Date,
		"filesize":      t.Filesize,
		"description":   util.SafeText(t.Description),
		"comments":      t.Comments,
		"category":      t.Category,
		"sub_category":  t.SubCategory,
		"downloads":     t.Downloads,
		"uploader_id":   t.UploaderID,
		"uploader_name": util.SafeText(t.Uploader.Username),
		"website_link":  util.Safe(t.WebsiteLink),
		"magnet":        template.URL(util.InfoHashToMagnet(strings.TrimSpace(t.Hash), t.Name, config.Trackers...)),
		"torrent":       util.Safe(torrentLink),
		"seeders":       t.Seeders,
		"leechers":      t.Leechers,
		"completed":     t.Completed,
		"last_scrape":   t.LastScrape,
	}
	return json.Marshal(pubForm)
}

/* We need a JSON object instead of a Gorm structure because magnet URLs are
   not in the database and have to be generated dynamically */

type ApiResultJSON struct {
	Torrents         []Torrent `json:"torrents"`
	QueryRecordCount int       `json:"queryRecordCount"`
	TotalRecordCount int       `json:"totalRecordCount"`
}

// TorrentObtainer is a function that obtains torrents somehow
type TorrentObtainer func() ([]Torrent, error)
