package commentService

import (
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
)

func GetAllComments(limit, offset uint32) (comments []model.Comment, err error) {
	comments, err = db.Impl.GetCommentsWhere(&common.CommentParam{
		Limit:  limit,
		Offset: offset,
	})
	return
}
func GetCommentsWhere(param *common.CommentParam) (comments []model.Comment, err error) {
	comments, err = db.Impl.GetCommentsWhere(param)
	return
}
