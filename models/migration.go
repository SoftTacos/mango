package models

import "time"

// represents a row in the db_version table

type MigrationDB struct {
	ID            uint
	FileID        uint   // integer before first underscore of filename
	Name          string // everything after the first underscore
	NextFileID    uint   // FileID of the file that must be applied after this migration is run
	OrderApplied  uint   //
	Applied       bool   //
	LastAppliedAt *time.Time
	QueryUp       []byte // query to upgrade version
	QueryDown     []byte // query to downgrade version
}

func NewMigration() Migration {
	migDB := &MigrationDB{}
	return Migration{
		MigrationDB: migDB,
		Query:       &migDB.QueryUp,
	}
}

type Migration struct {
	*MigrationDB
	NextMigration *Migration
	Query         *[]byte // todo: change to value
}
