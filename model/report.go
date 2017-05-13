package model

import (
	"time"
)

type TorrentReport struct {
	ID          uint32    `json:"id"`
	Description string    `json:"description"`
	TorrentID   uint32    `json:"torrent_id"`
	UserID      uint32    `json:"reported_by"`
	CreatedAt   time.Time `json:"reported_at"`
	Type        string    `json:"report_type"`
}
