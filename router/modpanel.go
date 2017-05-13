package router

// XXX: I hate this file so god damn much

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
	"github.com/ewhal/nyaa/service/comment"
	"github.com/ewhal/nyaa/service/report"
	"github.com/ewhal/nyaa/service/user"
	form "github.com/ewhal/nyaa/service/user/form"
	"github.com/ewhal/nyaa/service/user/permission"
	"github.com/ewhal/nyaa/util"
	"github.com/ewhal/nyaa/util/languages"
	"github.com/ewhal/nyaa/util/log"
	"github.com/ewhal/nyaa/util/search"
	"github.com/gorilla/mux"
)

func IndexModPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if userPermission.HasAdmin(currentUser) {
		limit := uint32(10)
		offset := uint32(0)
		torrents, err := db.Impl.GetTorrentsWhere(&common.TorrentParam{
			Max:    limit,
			Offset: offset,
		})
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}
		users, err := userService.RetrieveUsersForAdmin(limit, offset)
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}
		comments, err := commentService.GetAllComments(limit, offset)
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}

		torrentReports, err := reportService.GetTorrentReportsWhere(&common.ReportParam{
			Limit:  limit,
			Offset: offset,
		})

		languages.SetTranslationFromRequest(panelIndex, r, "en-us")
		htv := PanelIndexVbs{torrents, torrentReports, users, comments, NewSearchForm(), currentUser, r.URL}
		_ = panelIndex.ExecuteTemplate(w, "admin_index.html", htv)
	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}
}

func TorrentsListPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if userPermission.HasAdmin(currentUser) {
		vars := mux.Vars(r)
		page := vars["page"]

		var err error
		pagenum := uint32(0)
		if page != "" {
			p, err := strconv.Atoi(page)
			if err != nil {
				util.SendError(w, err, http.StatusInternalServerError)
				return
			}
			pagenum = uint32(p)
		}
		if pagenum == 0 {
			pagenum = 1
		}
		offset := uint32(100)

		searchParam, torrents, err := search.SearchByQuery(r, pagenum)
		searchForm := SearchForm{
			Torrent:            searchParam,
			HideAdvancedSearch: false,
		}

		languages.SetTranslationFromRequest(panelTorrentList, r, "en-us")
		htv := PanelTorrentListVbs{torrents, searchForm, Navigation{searchParam.Max, pagenum, "mod_tlist_page"}, currentUser, r.URL}
		err = panelTorrentList.ExecuteTemplate(w, "admin_index.html", htv)
		log.CheckError(err)
	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}
}

func TorrentReportListPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if userPermission.HasAdmin(currentUser) {
		vars := mux.Vars(r)
		page, err := strconv.ParseUint(vars["page"], 10, 32)
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}
		pagenum := uint32(page)
		if pagenum == 0 {
			pagenum = 1
		}
		limit := uint32(100)

		var reports []model.TorrentReport
		reports, err = reportService.GetTorrentReportsWhere(&common.ReportParam{
			Limit:  limit,
			Offset: uint32((pagenum - 1) * limit),
		})
		languages.SetTranslationFromRequest(panelTorrentReportList, r, "en-us")
		htv := PanelTorrentReportListVbs{reports, NewSearchForm(), Navigation{limit, pagenum, "mod_trlist_page"}, currentUser, r.URL}
		err = panelTorrentReportList.ExecuteTemplate(w, "admin_index.html", htv)
		log.CheckError(err)
	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}
}

func UsersListPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if userPermission.HasAdmin(currentUser) {
		vars := mux.Vars(r)
		page, err := strconv.ParseUint(vars["page"], 10, 32)
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}

		pagenum := uint32(page)
		if pagenum == 0 {
			pagenum = 1
		}
		limit := uint32(100)
		var users []model.User
		users, err = userService.RetrieveUsersForAdmin(limit, (pagenum-1)*limit)
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}
		languages.SetTranslationFromRequest(panelUserList, r, "en-us")
		htv := PanelUserListVbs{users, NewSearchForm(), Navigation{limit, pagenum, "mod_ulist_page"}, currentUser, r.URL}
		err = panelUserList.ExecuteTemplate(w, "admin_index.html", htv)
		log.CheckError(err)
	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}
}

func CommentsListPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if userPermission.HasAdmin(currentUser) {
		vars := mux.Vars(r)
		page, err := strconv.ParseUint(vars["page"], 10, 32)
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}

		pagenum := uint32(page)
		if pagenum == 0 {
			pagenum = 1
		}
		limit := uint32(100)

		var userid uint64
		userid, err = strconv.ParseUint(r.URL.Query().Get("userid"), 10, 32)
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}

		var comments []model.Comment
		comments, err = commentService.GetCommentsWhere(&common.CommentParam{
			Limit:  limit,
			Offset: uint32(pagenum-1) * limit,
			UserID: uint32(userid),
		})
		languages.SetTranslationFromRequest(panelCommentList, r, "en-us")
		htv := PanelCommentListVbs{comments, NewSearchForm(), Navigation{limit, pagenum, "mod_clist_page"}, currentUser, r.URL}
		err = panelCommentList.ExecuteTemplate(w, "admin_index.html", htv)
		log.CheckError(err)
	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}

}

func TorrentEditModPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if userPermission.HasAdmin(currentUser) {
		id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}
		var torrents []model.Torrent
		torrents, err = db.Impl.GetTorrentsWhere(&common.TorrentParam{
			TorrentID: uint32(id),
		})
		if err != nil {
			util.SendError(w, err, http.StatusInternalServerError)
			return
		}
		if len(torrents) == 0 {
			http.NotFound(w, r)
			return
		}
		torrent := torrents[0]
		languages.SetTranslationFromRequest(panelTorrentEd, r, "en-us")

		uploadForm := NewUploadForm()
		uploadForm.Name = torrent.Name
		uploadForm.Category = common.Category{
			Main: uint8(torrent.Category),
			Sub:  uint8(torrent.SubCategory),
		}.String()
		uploadForm.Status = torrent.Status
		uploadForm.Description = torrent.Description
		htv := PanelTorrentEdVbs{uploadForm, NewSearchForm(), currentUser, form.NewErrors(), form.NewInfos(), r.URL}
		err = panelTorrentEd.ExecuteTemplate(w, "admin_index.html", htv)
		log.CheckError(err)

	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}

}

func TorrentPostEditModPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if !userPermission.HasAdmin(currentUser) {
		http.Error(w, "admins only", http.StatusForbidden)
		return
	}
	var uploadForm UploadForm
	id, e := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)
	if e != nil {
		util.SendError(w, e, http.StatusInternalServerError)
		return
	}
	err := form.NewErrors()
	infos := form.NewInfos()
	torrents, er := db.Impl.GetTorrentsWhere(&common.TorrentParam{TorrentID: uint32(id)})
	if er == nil {
		if len(torrents) == 0 {
			http.NotFound(w, r)
			return
		}
		errUp := uploadForm.ExtractEditInfo(r)
		if errUp != nil {
			err["errors"] = append(err["errors"], "Failed to update torrent!")
		}
		if len(err) == 0 {
			torrent := torrents[0]
			// update some (but not all!) values
			torrent.Name = uploadForm.Name
			torrent.Category = uploadForm.CategoryID
			torrent.SubCategory = uploadForm.SubCategoryID
			torrent.Status = uploadForm.Status
			torrent.Description = uploadForm.Description
			e := db.Impl.UpsertTorrent(&torrent)
			if e == nil {
				infos["infos"] = append(infos["infos"], "Torrent details updated.")
			} else {
				err["errors"] = append(err["errors"], e.Error())
			}

		}
		languages.SetTranslationFromRequest(panelTorrentEd, r, "en-us")
		htv := PanelTorrentEdVbs{uploadForm, NewSearchForm(), currentUser, err, infos, r.URL}
		_ = panelTorrentEd.ExecuteTemplate(w, "admin_index.html", htv)
	} else {
		http.Error(w, er.Error(), http.StatusInternalServerError)
	}
}

func CommentDeleteModPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)

	if userPermission.HasAdmin(currentUser) {
		id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = form.NewErrors()
		_, _ = userService.DeleteComment(uint32(id))
		url, _ := Router.Get("mod_clist").URL()
		http.Redirect(w, r, url.String()+"?deleted", http.StatusSeeOther)
	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}
}

func TorrentDeleteModPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if userPermission.HasAdmin(currentUser) {
		id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		db.Impl.DeleteTorrentsWhere(&common.TorrentParam{
			TorrentID: uint32(id),
		})

		db.Impl.DeleteTorrentReportsWhere(&common.ReportParam{
			TorrentID: uint32(id),
		})

		//delete reports of torrent
		whereParams := serviceBase.CreateWhereParams("torrent_id = ?", id)
		reports, _, _ := reportService.GetTorrentReportsOrderBy(&whereParams, "", 0, 0)
		for _, report := range reports {
			reportService.DeleteTorrentReport(report.ID)
		}
		url, _ := Router.Get("mod_tlist").URL()
		http.Redirect(w, r, url.String()+"?deleted", http.StatusSeeOther)
	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}
}

func TorrentReportDeleteModPanel(w http.ResponseWriter, r *http.Request) {
	currentUser := GetUser(r)
	if userPermission.HasAdmin(currentUser) {
		id := r.URL.Query().Get("id")
		fmt.Println(id)
		idNum, _ := strconv.ParseUint(id, 10, 64)
		_ = form.NewErrors()
		_, _ = reportService.DeleteTorrentReport(uint(idNum))

		url, _ := Router.Get("mod_trlist").URL()
		http.Redirect(w, r, url.String()+"?deleted", http.StatusSeeOther)
	} else {
		http.Error(w, "admins only", http.StatusForbidden)
	}
}
