package model

import (
	"time"
)

// TODO Add field to specify kind of reports
// TODO Add CreatedAt field
// INFO User can be null (anonymous reports)
// FIXME  can't preload field Torrents for model.TorrentReport
type TorrentReport struct {
	ID          uint
	Description string
	TorrentID   uint
	UserID      uint

	CreatedAt time.Time

	Type string

	Torrent *Torrent
	User    *User
}

type TorrentReportJson struct {
	ID          uint         `json:"id"`
	Description string       `json:"description"`
	Torrent     *TorrentJSON `json:"torrent"`
	User        *UserJSON    `json:"user"`
}

/* Model Conversion to Json */

func (report *TorrentReport) ToJson() *TorrentReportJson {
	json := &TorrentReportJson{report.ID, report.Description, report.Torrent.ToJSON(), report.User.ToJSON()}
	return json
}

func TorrentReportsToJSON(reports []*TorrentReport) []*TorrentReportJson {
	json := make([]*TorrentReportJson, len(reports))
	for i := range reports {
		json[i] = reports[i].ToJson()
	}
	return json
}
