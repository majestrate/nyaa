package postgres

import (
	"database/sql"

	"github.com/ewhal/nyaa/model"
)

func (db *Database) UserFollows(a, b uint32) (follows bool, err error) {
	err = db.queryWithPrepared(queryUserFollows, func(rows *sql.Rows) error {
		follows = true
		return nil
	}, a, b)
	return
}

func (db *Database) getUserByParam(name string, p interface{}) (user model.User, has bool, err error) {
	err = db.queryWithPrepared(name, func(rows *sql.Rows) error {
		rows.Next()
		scanUserColumnsFull(rows, &user)
		has = true
		return nil
	}, p)
	return
}

func (db *Database) GetUserByApiToken(token string) (user model.User, has bool, err error) {
	user, has, err = db.getUserByParam(queryGetUserByApiToken, token)
	return
}

func (db *Database) GetUsersByEmail(email string) (users []model.User, err error) {
	err = db.queryWithPrepared(queryGetUserByEmail, func(rows *sql.Rows) error {
		for rows.Next() {
			var user model.User
			scanUserColumnsFull(rows, &user)
			users = append(users, user)
		}
		return nil
	}, email)
	return
}

func (db *Database) GetUserByName(name string) (user model.User, has bool, err error) {
	user, has, err = db.getUserByParam(queryGetUserByName, name)
	return
}

func (db *Database) GetUserByID(id uint32) (user model.User, has bool, err error) {
	user, has, err = db.getUserByParam(queryGetUserByID, id)
	return
}

func (db *Database) InsertUser(u *model.User) (err error) {
	_, err = db.getPrepared(queryInsertUser).Exec(u.Username, u.Password, u.Email, u.Status, u.CreatedAt, u.UpdatedAt, u.LastLoginAt, u.LastLoginIP, u.Token, u.TokenExpiration, u.Language, u.MD5)
	return
}

func (db *Database) UpdateUser(u *model.User) (err error) {
	_, err = db.getPrepared(queryUpdateUser).Exec(u.ID, u.Username, u.Password, u.Email, u.Status, u.UpdatedAt, u.LastLoginAt, u.LastLoginIP, u.Token, u.TokenExpiration, u.Language, u.MD5)
	return
}

func (db *Database) DeleteUserByID(id uint32) (err error) {
	_, err = db.getPrepared(queryDeleteUserByID).Exec(id)
	return
}
