package search

import (
	"net/http"
	"strconv"

	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
)

func Configure(conf *config.SearchConfig) (err error) {

	return
}

func SearchByQuery(r *http.Request, pagenum uint32) (search common.SearchParam, tor []model.Torrent, count int, err error) {
	search, tor, count, err = searchByQuery(r, pagenum, true)
	return
}

func SearchByQueryNoCount(r *http.Request, pagenum uint32) (search common.SearchParam, tor []model.Torrent, err error) {
	search, tor, _, err = searchByQuery(r, pagenum, false)
	return
}

func getRequestSearchParam(r *http.Request, pagenum uint32) (search common.SearchParam) {

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

	return
}

func searchByQuery(r *http.Request, pagenum uint32, countAll bool) (
	search common.SearchParam, torrents []model.Torrent, count int, err error,
) {
	search = getRequestSearchParam(r, pagenum)
	torrents, err = db.Impl.GetTorrentsWhere(&search)
	return
}
