package model

import (
	"time"
)

type User struct {
	ID              uint32
	Username        string
	Password        string
	Email           string
	Status          int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Token           string
	TokenExpiration time.Time
	Language        string

	// TODO: move this to PublicUser
	LikingCount int    `json:"likingCount"`
	LikedCount  int    `json:"likedCount"`
	Likings     []User // Don't work `gorm:"foreignkey:user_id;associationforeignkey:follower_id;many2many:user_follows"`
	Liked       []User // Don't work `gorm:"foreignkey:follower_id;associationforeignkey:user_id;many2many:user_follows"`

	MD5      string `json:"md5"` // Hash of email address, used for Gravatar
	Torrents []Torrent

	LastLoginAt time.Time
	LastLoginIP string
}

type UserJSON struct {
	ID          uint32 `json:"user_id"`
	Username    string `json:"username"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at"`
	LikingCount int    `json:"liking_count"`
	LikedCount  int    `json:"liked_count"`
}

// Returns the total size of memory recursively allocated for this struct
func (u User) Size() (s int) {
	s += 4 + // ints
		6*2 + // string pointers
		4*3 + //time.Time
		3*2 + // arrays
		// string arrays
		len(u.Username) + len(u.Password) + len(u.Email) + len(u.Token) + len(u.MD5) + len(u.Language)
	s *= 8

	// Ignoring foreign key users. Fuck them.

	return
}

type PublicUser struct {
	User *User
}

// different users following eachother
type UserFollows struct {
	UserID     uint32
	FollowerID uint32
}

type UserUploadsOld struct {
	Username  string
	TorrentId uint
}

func (u *User) ToJSON() *UserJSON {
	json := &UserJSON{
		ID:          u.ID,
		Username:    u.Username,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		LikingCount: u.LikingCount,
		LikedCount:  u.LikedCount,
	}
	return json
}
