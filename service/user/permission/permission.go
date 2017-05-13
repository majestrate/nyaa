package userPermission

import (
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
	"github.com/ewhal/nyaa/util/log"
)

// HasAdmin checks that user has an admin permission.
func HasAdmin(user *model.User) bool {
	return user.Status == 2
}

// CurrentOrAdmin check that user has admin permission or user is the current user.
func CurrentOrAdmin(user *model.User, userID uint32) bool {
	log.Debugf("user.ID == userID %d %d %s", user.ID, userID, user.ID == userID)
	return (HasAdmin(user) || user.ID == userID)
}

// CurrentUserIdentical check that userID is same as current user's ID.
// TODO: Inline this
func CurrentUserIdentical(user *model.User, userID uint32) bool {
	return user.ID == userID
}

func GetRole(user *model.User) (str string) {
	switch user.Status {
	case -1:
		str = "Banned"
	case 1:
		str = "Trusted Member"
	case 2:
		str = "Moderator"
	case 0:
	default:
		str = "Member"
	}
	return
}

func IsFollower(user *model.User, currentUser *model.User) (follows bool, err error) {
	follows, err = db.Impl.UserFollows(currentUser.ID, user.ID)
	return
}
