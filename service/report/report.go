package reportService

import (
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
	"net/http"
)

func DeleteTorrentReport(id uint32) (err error, status int) {
	err = db.Impl.DeleteTorrentReportByID(id)
	if err == nil {
		status = http.StatusOK
	} else {
		status = http.StatusInternalServerError
	}
	return
}

func GetTorrentReportsWhere(parameters *common.ReportParam) ([]model.TorrentReport, error) {
	return db.Impl.GetTorrentReportsWhere(parameters)
}
