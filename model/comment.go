package model

import (
	"time"
)

type Comment struct {
	ID        uint
	TorrentID uint
	UserID    uint
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Torrent *Torrent
	User    *User
}

// Returns the total size of memory recursively allocated for this struct
func (c Comment) Size() int {
	return (3 + 3*3 + 2 + 2 + len(c.Content)) * 8
}

type OldComment struct {
	TorrentID uint
	Username  string
	Content   string
	Date      time.Time

	Torrent *Torrent
}

// Returns the total size of memory recursively allocated for this struct
func (c OldComment) Size() int {
	return (1 + 2*2 + len(c.Username) + len(c.Content) + 3 + 1) * 8
}
