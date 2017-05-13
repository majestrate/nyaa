package postgres

import (
	"github.com/ewhal/nyaa/model"
)

func (db *Database) InsertComment(comment *model.Comment) (err error) {
	_, err = db.getPrepared(queryInsertComment).Exec(comment.ID, comment.TorrentID, comment.Content, comment.CreatedAt)
	return
}
