package models

import "time"

// represents a row in the db_version table

type MigrationDB struct {
	ID              uint
	FileID          uint   // integer before first underscore of filename
	Name            string // everything after the first underscore
	RequiredFileIDs []uint // FileIDs of the migrations that must be run before this migration is applied
	OrderApplied    uint   //
	Applied         bool   //
	LastAppliedAt   *time.Time
	QueryUp         []byte // query to upgrade version
	QueryDown       []byte // query to downgrade version
}

func NewMigration() Migration {
	migDB := &MigrationDB{
		RequiredFileIDs: []uint{},
	}
	return Migration{
		MigrationDB: migDB,
		Query:       true,
	}
}

type Migration struct {
	*MigrationDB
	NextMigration *Migration
	Query         bool
}
