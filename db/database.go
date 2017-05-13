package db

import (
	"github.com/ewhal/nyaa/common"
	"github.com/ewhal/nyaa/config"
	"github.com/ewhal/nyaa/db/postgres"
	//"github.com/ewhal/nyaa/db/sqlite"
	"github.com/ewhal/nyaa/model"
	"github.com/ewhal/nyaa/util/log"

	"database/sql"
	"errors"
)

// Database obstraction layer
type Database interface {

	// Initialize internal state
	Init() error

	// return true if we need to call MigrateNext again
	NeedsMigrate() (bool, error)
	// migrate to next database revision
	MigrateNext() error

	// get all torrents with limit and offset
	GetAllTorrents(offset, limit uint32) ([]model.Torrent, error)
	// get all torrents with where params
	GetTorrentsWhere(param *common.SearchParam) ([]model.Torrent, error)
	// get torrent by id
	GetTorrentByID(id uint32) (model.Torrent, bool, error)
	// insert new comment
	InsertComment(comment *model.Comment) error
	// new torrent report
	InsertTorrentReport(report *model.TorrentReport) error

	// check if user A follows B (by id)
	UserFollows(a, b uint32) (bool, error)

	DeleteTorrentReportByID(id uint32) error

	GetTorrentReportsWhere(param *common.ReportParam) ([]model.TorrentReport, error)

	// bulk record scrape events in 1 transaction
	RecordScrapes(scrapes []common.ScrapeResult) error

	// get user model by api token
	GetUserByApiToken(token string) (model.User, bool, error)

	// get users by email
	GetUsersByEmail(email string) ([]model.User, error)

	// get user by username
	GetUserByName(name string) (model.User, bool, error)

	// get user by ID
	GetUserByID(id uint32) (model.User, bool, error)

	// insert new user
	InsertUser(u *model.User) error

	// update existing user info
	UpdateUser(u *model.User) error

	// delete a user by ID
	DeleteUserByID(id uint32) error

	// DO NOT USE ME kthnx
	Query(query string, params ...interface{}) (*sql.Rows, error)
}

var ErrInvalidDatabaseDialect = errors.New("invalid database dialect")
var ErrSqliteSucksAss = errors.New("sqlite3 sucks ass so it's not supported yet")

var Impl Database

func Configure(conf *config.Config) (err error) {
	switch conf.DBType {
	case "postgres":
		Impl, err = postgres.New(conf.DBParams)
		break
	case "sqlite3":
		err = ErrSqliteSucksAss
		// Impl, err = sqlite.New(conf.DBParams)
		break
	default:
		err = ErrInvalidDatabaseDialect
	}
	if err == nil {
		log.Infof("Init %s database", conf.DBType)
		err = Impl.Init()
	}
	return
}

// Migrate migrates the database to latest revision, call after Configure
func Migrate() (err error) {
	next := true
	for err == nil && next {
		next, err = Impl.NeedsMigrate()
		if next {
			err = Impl.MigrateNext()
		}

	}
	return
}
