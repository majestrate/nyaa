package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
	"github.com/ewhal/nyaa/service/api"
	"github.com/ewhal/nyaa/util"
	"github.com/ewhal/nyaa/util/log"
	"github.com/gorilla/mux"
)

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page := vars["page"]
	whereParams := common.TorrentParam{}
	req := apiService.TorrentsRequest{}

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		d := json.NewDecoder(r.Body)
		if err := d.Decode(&req); err != nil {
			util.SendError(w, err, 502)
		}

		if req.MaxPerPage == 0 {
			req.MaxPerPage = 50
		}
		if req.Page == 0 {
			req.Page = 1
		}

		whereParams = req.ToParams()
	} else {
		var err error
		maxString := r.URL.Query().Get("max")
		if maxString != "" {
			var i uint64
			i, err = strconv.ParseUint(maxString, 10, 32)
			if !log.CheckError(err) {
				req.MaxPerPage = 50 // default Value maxPerPage
			} else {
				req.MaxPerPage = uint32(i)
			}
		}

		req.Page = 1
		if page != "" {
			var i uint64
			i, err = strconv.ParseUint(page, 10, 32)
			if !log.CheckError(err) {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			req.Page = uint32(i)

		}
	}

	whereParams.Max, whereParams.Offset = req.MaxPerPage, req.MaxPerPage*(req.Page-1)

	whereParams.Full = true

	torrents, err := db.Impl.GetTorrentsWhere(&whereParams)
	if err != nil {
		util.SendError(w, err, 400)
		return
	}

	b := model.ApiResultJSON{
		Torrents: torrents,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ApiViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	var result interface{}
	w.Header().Set("Content-Type", "application/json")
	if err == nil {
		torrents, err := db.Impl.GetTorrentsWhere(&common.TorrentParam{
			TorrentID: uint32(id),
		})
		if err == nil {
			if len(torrents) > 0 {
				result = torrents[0]
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			result = map[string]interface{}{"error": err.Error()}
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		result = map[string]interface{}{"error": err.Error()}
	}
	json.NewEncoder(w).Encode(result)
}

func apiUpsertTorrentHandler(w http.ResponseWriter, r *http.Request, fresh bool) {
	var result interface{}
	var status int
	if config.UploadsDisabled {
		status = http.StatusForbidden
		result = map[string]interface{}{"error": "uploads disabled"}

	} else {

		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" {
			token := r.Header.Get("Authorization")
			users, err := db.Impl.GetUsersWhere(&common.UserParam{
				ApiToken: token,
			})
			if err == nil {
				if len(users) == 0 {
					result = map[string]interface{}{"error": apiService.ErrApiKey.Error()}
					status = http.StatusForbidden
				} else {
					user := users[0]
					if user.ID == 0 {
						result = map[string]interface{}{"error": apiService.ErrApiKey.Error()}
						status = http.StatusForbidden
					} else {

						defer r.Body.Close()

						if err == nil {
							var torrent model.Torrent
							dec := json.NewDecoder(r.Body)
							upload := apiService.TorrentRequest{}
							update := apiService.UpdateRequest{}
							if fresh {
								err = dec.Decode(&upload)
								if err == nil {
									err, status = upload.ValidateUpload()
									if err == nil {
										torrent = model.Torrent{
											Name:        upload.Name,
											Category:    upload.Category,
											SubCategory: upload.SubCategory,
											Status:      1,
											Hash:        upload.Hash,
											Date:        time.Now(),
											Filesize:    0, //?
											Description: upload.Description,
											UploaderID:  user.ID,
											Uploader:    &user,
										}
									}
								} else {
									status = http.StatusBadRequest
								}
							} else {
								err = dec.Decode(&update)
								if err == nil {
									torrent.Name = update.Update.Name
									torrent.Category = update.Update.Category
									torrent.SubCategory = update.Update.SubCategory
									torrent.Description = update.Update.Description
									torrent.ID = update.ID
								} else {
									status = http.StatusBadRequest
								}
							}
							if err == nil {
								err = db.Impl.UpsertTorrent(&torrent)
								if err == nil {
									status = http.StatusCreated
									result = &torrent
								} else {
									status = http.StatusInternalServerError
									result = map[string]interface{}{
										"error": err.Error(),
									}
								}
							} else {
								result = map[string]interface{}{
									"error": err.Error(),
								}
							}

						} else {
							status = http.StatusBadRequest
							result = map[string]interface{}{
								"error": err.Error(),
							}
						}
					}
				}
			} else {
				status = http.StatusInternalServerError
				result = map[string]interface{}{
					"error": err.Error(),
				}
			}
		} else {
			status = http.StatusUnsupportedMediaType
			result = map[string]interface{}{
				"error": fmt.Sprintf("bad content type: %s", contentType),
			}
		}
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(result)
}

func ApiUpdateHandler(w http.ResponseWriter, r *http.Request) {
	apiUpsertTorrentHandler(w, r, false)
}

func ApiUploadHandler(w http.ResponseWriter, r *http.Request) {
	apiUpsertTorrentHandler(w, r, true)
}
