package userService

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/db"
	"github.com/ewhal/nyaa/model"
	formStruct "github.com/ewhal/nyaa/service/user/form"
	"github.com/ewhal/nyaa/service/user/permission"
	"github.com/ewhal/nyaa/util/crypto"
	"github.com/ewhal/nyaa/util/log"
	"github.com/ewhal/nyaa/util/modelHelper"
	"github.com/ewhal/nyaa/util/timeHelper"
	"golang.org/x/crypto/bcrypt"
)

// SuggestUsername suggest user's name if user's name already occupied.
func SuggestUsername(username string) (name string) {
	// TODO: make this less shit
	tries := 10
	for tries > 0 {
		num := rand.Intn(9005)
		name = fmt.Sprintf("%s_%d", username, num)
		users, err := db.Impl.GetUsersWhere(&common.UserParam{
			Name: name,
		})
		if len(users) == 0 {
			return
		} else if err != nil {
			break
		}
		tries--

	}
	name = "gayballs" // lolz
	return
}

func CheckEmail(email string) (bool, error) {
	if len(email) > 0 {
		users, err := db.Impl.GetUsersWhere(&common.UserParam{})
		if err == nil {
			return len(users) > 0, nil
		}
		return false, err
	}
	return false, nil
}

// CreateUserFromForm creates a user from a registration form.
func CreateUserFromForm(registrationForm formStruct.RegistrationForm) (user model.User, err error) {
	log.Debugf("registrationForm %+v\n", registrationForm)
	modelHelper.AssignValue(&user, &registrationForm)
	if user.Email == "" {
		user.MD5 = ""
	} else {
		// Despite the email not being verified yet we calculate this for convenience reasons
		user.MD5, err = crypto.GenerateMD5Hash(user.Email)
		if err != nil {
			return
		}
	}
	token, err := crypto.GenerateRandomToken32()
	if err != nil {
		return
	}
	user.Email = "" // unset email because it will be verified later

	user.Token = token
	user.TokenExpiration = timeHelper.FewDaysLater(config.AuthTokenExpirationDay)
	user.CreatedAt = time.Now()
	err = db.Impl.InsertUser(&user)
	return
}

// CreateUser creates a user.
func CreateUser(w http.ResponseWriter, r *http.Request) (int, error) {
	var user model.User
	var registrationForm formStruct.RegistrationForm
	var status int

	modelHelper.BindValueForm(&registrationForm, r)
	usernameCandidate := SuggestUsername(registrationForm.Username)
	if usernameCandidate != registrationForm.Username {
		return http.StatusInternalServerError, fmt.Errorf("Username already taken, you can choose: %s", usernameCandidate)
	}

	if registrationForm.Email == "" {
		return http.StatusInternalServerError, common.ErrBadEmail
	} else {
		check, err := CheckEmail(registrationForm.Email)
		if check {
			return http.StatusInternalServerError, common.ErrUserExists
		} else if err != nil {
			return http.StatusInternalServerError, err
		}
	}
	password, err := bcrypt.GenerateFromPassword([]byte(registrationForm.Password), 10)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	registrationForm.Password = string(password)
	user, err = CreateUserFromForm(registrationForm)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	SendVerificationToUser(user, registrationForm.Email)

	status, err = RegisterHandler(w, r)
	return status, err
}

// RetrieveUser retrieves a user by ID
func RetrieveUser(r *http.Request, id uint32) (model.PublicUser, bool, uint32, int, error) {
	var pubuser model.PublicUser
	var currentUserID uint32
	var isAuthor bool
	users, err := db.Impl.GetUsersWhere(&common.UserParam{
		ID: id,
	})
	if err != nil {
		return pubuser, false, 0, http.StatusInternalServerError, err
	}
	if len(users) == 0 {
		return pubuser, isAuthor, currentUserID, http.StatusNotFound, common.ErrNoSuchUser
	}
	user := users[0]
	currentUser, err := CurrentUser(r)
	if err == nil {
		currentUserID = currentUser.ID
		isAuthor = currentUser.ID == user.ID
	}

	return user.ToPublic(), isAuthor, currentUserID, http.StatusOK, nil
}

// RetrieveUsers retrieves users.
func RetrieveUsers() []model.PublicUser {
	var users []model.User
	var userArr []model.PublicUser
	for _, user := range users {
		userArr = append(userArr, user.ToPublic())
	}
	return userArr
}

// UpdateUserCore updates a user. (Applying the modifed data of user).
func UpdateUserCore(user *model.User) (int, error) {
	if user.Email == "" {
		user.MD5 = ""
	} else {
		var err error
		user.MD5, err = crypto.GenerateMD5Hash(user.Email)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	token, err := crypto.GenerateRandomToken32()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	user.Token = token
	user.TokenExpiration = timeHelper.FewDaysLater(config.AuthTokenExpirationDay)
	user.UpdatedAt = time.Now()
	err = db.Impl.UpdateUser(user)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// UpdateUser updates a user.
func UpdateUser(w http.ResponseWriter, form *formStruct.UserForm, currentUser *model.User, id uint32) (user model.User, status int, err error) {
	var users []model.User
	users, err = db.Impl.GetUsersWhere(&common.UserParam{
		ID: id,
	})
	if len(users) == 0 {
		if err == nil {
			status = http.StatusNotFound
			err = common.ErrNoSuchUser
		} else {
			status = http.StatusInternalServerError
		}
		return
	}
	user = users[0]
	if form.Password != "" {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.CurrentPassword))
		if err != nil && !userPermission.HasAdmin(currentUser) {
			log.Error("Password Incorrect.")
			return user, http.StatusInternalServerError, common.ErrBadLogin
		}
		newPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), 10)
		if err != nil {
			return user, http.StatusInternalServerError, err
		}
		form.Password = string(newPassword)
	} else { // Then no change of password
		form.Password = user.Password
	}
	if !userPermission.HasAdmin(currentUser) { // We don't want users to be able to modify some fields
		form.Status = user.Status
		form.Username = user.Username
	}
	if form.Email != user.Email {
		SendVerificationToUser(user, form.Email)
		form.Email = user.Email
	}
	modelHelper.AssignValue(&user, form)
	status, err = UpdateUserCore(&user)
	if err != nil {
		return
	}
	if userPermission.CurrentUserIdentical(currentUser, user.ID) {
		status, err = SetCookie(w, user.Token)
	}
	return
}

// DeleteUser deletes a user.
func DeleteUser(w http.ResponseWriter, currentUser *model.User, id uint32) (status int, err error) {
	status = http.StatusOK
	if currentUser.ID == 0 {
		err = common.ErrInsufficientPermission
		status = http.StatusForbidden
		return
	}
	if userPermission.CurrentOrAdmin(currentUser, id) {
		var affected uint32
		affected, err = db.Impl.DeleteUsersWhere(&common.UserParam{
			ID: id,
		})
		if err != nil {
			status = http.StatusInternalServerError
			return
		}
		if affected == 0 {
			err = common.ErrNoSuchUser
			status = http.StatusInternalServerError
		} else if userPermission.CurrentUserIdentical(currentUser, id) {
			status, err = ClearCookie(w)
		}
	} else {
		err = common.ErrInsufficientPermission
		status = http.StatusForbidden
	}
	return
}

// RetrieveCurrentUser retrieves a current user.
func RetrieveCurrentUser(r *http.Request) (user model.User, status int, err error) {
	user, err = CurrentUser(r)
	if err == nil {
		status = http.StatusOK
	} else {
		status = http.StatusInternalServerError
	}
	return
}

// RetrieveUsersByEmail retrieves users by an email
func RetrieveUsersByEmail(email string) (pubusers []model.PublicUser, err error) {
	var users []model.User
	users, err = db.Impl.GetUsersWhere(&common.UserParam{
		Email: email,
	})
	if err == nil && len(users) > 0 {
		pubusers = make([]model.PublicUser, len(users))
		for idx := range users {
			pubusers[idx] = users[idx].ToPublic()
		}
	}
	return
}

// RetrieveUserByUsername retrieves a user by username.
func RetrieveUserByUsername(name string) (pubuser model.PublicUser, username string, status int, err error) {
	var users []model.User
	username = name
	users, err = db.Impl.GetUsersWhere(&common.UserParam{
		Name: name,
	})
	if len(users) > 0 {
		pubuser = users[0].ToPublic()
		status = http.StatusOK
	} else if err == nil {
		err = common.ErrNoSuchUser
		status = http.StatusNotFound
	} else {
		status = http.StatusInternalServerError
	}
	return
}

// RetrieveUserForAdmin retrieves a user for an administrator.
func RetrieveUserForAdmin(id uint32) (user model.User, status int, err error) {
	var users []model.User
	users, err = db.Impl.GetUsersWhere(&common.UserParam{
		Full: true,
		ID:   id,
	})
	if len(users) > 0 {
		status = http.StatusOK
		user = users[0]
	} else if err == nil {
		status = http.StatusNotFound
		err = common.ErrNoSuchUser
	}
	return
}

// RetrieveUsersForAdmin retrieves users for an administrator.
func RetrieveUsersForAdmin(limit uint32, offset uint32) (users []model.User, err error) {
	users, err = db.Impl.GetUsersWhere(&common.UserParam{
		Full:   true,
		Max:    limit,
		Offset: offset,
	})
	return
}

// CreateUserAuthentication creates user authentication.
func CreateUserAuthentication(w http.ResponseWriter, r *http.Request) (status int, err error) {
	var form formStruct.LoginForm
	modelHelper.BindValueForm(&form, r)
	status, err = SetCookieHandler(w, form.Username, form.Password)
	return
}

func SetFollow(user *model.User, follower *model.User) (err error) {
	if follower.ID > 0 && user.ID > 0 {
		err = db.Impl.AddUserFollowing(user.ID, follower.ID)
	} else {
		err = common.ErrNoSuchEntry
	}
	return
}

func RemoveFollow(user *model.User, follower *model.User) (err error) {
	if follower.ID > 0 && user.ID > 0 {
		var deleted bool
		deleted, err = db.Impl.DeleteUserFollowing(user.ID, follower.ID)
		if err == nil && !deleted {
			err = common.ErrNotFollowing
		}
	} else {
		err = common.ErrNoSuchEntry
	}
	return
}

func DeleteComment(id uint32) (status int, err error) {
	var affected uint32
	affected, err = db.Impl.DeleteCommentsWhere(&common.CommentParam{
		CommentID: id,
	})
	if affected > 0 {
		status = http.StatusOK
	} else if err == nil {
		status = http.StatusNotFound
		err = common.ErrNoSuchComment
	} else {
		status = http.StatusInternalServerError
	}
	return
}
