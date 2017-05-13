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
	MD5             string
	LastLoginAt     time.Time
	LastLoginIP     string
	Uploads         []Torrent
	UsersWeLiked    []uint32
	UsersLikingMe   []uint32
}

type PublicUser struct {
	ID          uint32    `json:"user_id"`
	Username    string    `json:"username"`
	Status      int       `json:"status"`
	CreatedAt   string    `json:"created_at"`
	LikingCount int       `json:"liking_count"`
	LikedCount  int       `json:"liked_count"`
	MD5         string    `json:"md5"`
	Uploads     []Torrent `json:"uploads"`
}

func (u *User) ToPublic() PublicUser {
	return PublicUser{
		ID:          u.ID,
		Username:    u.Username,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		LikingCount: len(u.UsersWeLiked),
		LikedCount:  len(u.UsersLikingMe),
		MD5:         u.MD5,
		Uploads:     u.Uploads,
	}
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

// different users following eachother
type UserFollows struct {
	UserID     uint32
	FollowerID uint32
}

type UserUploadsOld struct {
	Username  string
	TorrentId uint
}
