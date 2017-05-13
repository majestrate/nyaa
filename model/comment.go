package model

import (
	"encoding/json"
	"github.com/ewhal/nyaa/util"
	"time"
)

type Comment struct {
	ID        uint32
	TorrentID uint32
	UserID    uint32
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (c *Comment) MarshalJSON() ([]byte, error) {
	pub := map[string]interface{}{
		"id":         c.ID,
		"torrent_id": c.TorrentID,
		"user_id":    c.UserID,
		"content":    util.MarkdownToHTML(c.Content),
		"created_at": c.CreatedAt,
		"updated_at": c.UpdatedAt,
	}
	return json.Marshal(pub)

}

// Returns the total size of memory recursively allocated for this struct
func (c Comment) Size() int {
	return (3 + 3*3 + 2 + 2 + len(c.Content)) * 8
}

type OldComment struct {
	TorrentID uint32
	Username  string
	Content   string
	Date      time.Time
}

// Returns the total size of memory recursively allocated for this struct
func (c OldComment) Size() int {
	return (1 + 2*2 + len(c.Username) + len(c.Content) + 3 + 1) * 8
}
