package search

import (
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ewhal/nyaa/cache"
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
)

func Configure(conf *config.SearchConfig) (err error) {

	return
}

func getRequestTorrentParam(r *http.Request, pagenum uint32) (search common.TorrentParam) {

	max, err := strconv.ParseUint(r.URL.Query().Get("max"), 10, 32)
	if err != nil {
		err = nil
		search.Max = 50
	}
	if search.Max > 300 {
		search.Max = 300
	} else {
		search.Max = uint32(max)
	}
	if pagenum == 0 {
		pagenum = 1
	}
	search.Offset = (pagenum - 1) * search.Max

	q := r.URL.Query().Get("q")
	var words []string
	split := strings.Fields(q)
	for idx := range split {
		firstRune, _ := utf8.DecodeRuneInString(split[idx])
		if len(split[idx]) == 1 && (unicode.IsPunct(firstRune) || unicode.IsControl(firstRune)) {
			continue
		}
		words = append(words, split[idx])
	}

	search.NameLike = strings.Join(split, ",")

	var userid uint64
	userid, _ = strconv.ParseUint(r.URL.Query().Get("userID"), 10, 32)
	search.UserID = uint32(userid)

	search.Category.Parse(r.URL.Query().Get("c"))
	search.Status.Parse(r.URL.Query().Get("s"))
	search.Sort.Parse(r.URL.Query().Get("sort"))

	var notNulls []string
	switch search.Sort {
	case common.Seeders:
		notNulls = append(notNulls, "seeders")
		break
	case common.Leechers:
		notNulls = append(notNulls, "leechers")
		break
	case common.Completed:
		notNulls = append(notNulls, "completed")
		break
	}

	search.NotNull = strings.Join(notNulls, ",")

	if r.URL.Query().Get("order") == "true" {
		search.Order = true
	}

	return
}

func SearchByQuery(r *http.Request, pagenum uint32) (search common.TorrentParam, torrents []model.Torrent, err error) {
	search = getRequestTorrentParam(r, pagenum)
	torrents, err = cache.Impl.GetTorrents(&search, func() ([]model.Torrent, error) {
		return db.Impl.GetTorrentsWhere(&search)
	})
	return
}
