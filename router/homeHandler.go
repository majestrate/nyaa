package router

import (
	"net/http"
	"strconv"

	"github.com/ewhal/nyaa/cache"
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
	"github.com/ewhal/nyaa/util"
	"github.com/ewhal/nyaa/util/languages"
	"github.com/ewhal/nyaa/util/log"
	"github.com/gorilla/mux"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page := vars["page"]

	// db params url
	var err error
	maxPerPage := uint32(50)
	maxString := r.URL.Query().Get("max")
	if maxString != "" {
		var p uint64
		p, err = strconv.ParseUint(maxString, 10, 32)
		if !log.CheckError(err) {
			maxPerPage = uint32(50) // default Value maxPerPage
		} else {
			maxPerPage = uint32(p)
		}
	}

	pagenum := uint32(0)
	if page != "" {
		var p uint64
		p, err = strconv.ParseUint(page, 10, 32)
		if !log.CheckError(err) {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}
		pagenum = uint32(p)
	}

	if pagenum == 0 {
		pagenum = 1
	}

	search := common.TorrentParam{
		Max:    uint32(maxPerPage),
		Offset: uint32(pagenum-1) * uint32(maxPerPage),
	}

	torrents, err := cache.Impl.GetTorrents(&search, func() ([]model.Torrent, error) {
		return db.Impl.GetTorrentsWhere(&search)
	})
	if err != nil {
		util.SendError(w, err, http.StatusInternalServerError)
		return
	}

	navigationTorrents := Navigation{maxPerPage, pagenum, "search_page"}

	languages.SetTranslationFromRequest(homeTemplate, r, "en-us")
	htv := HomeTemplateVariables{torrents, NewSearchForm(), navigationTorrents, GetUser(r), r.URL, mux.CurrentRoute(r)}

	err = homeTemplate.ExecuteTemplate(w, "index.html", htv)
	if err != nil {
		log.Errorf("HomeHandler(): %s", err)
	}
}
