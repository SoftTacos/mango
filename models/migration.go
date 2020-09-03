package models

import "time"

// represents a row in the db_version table

type MigrationDB struct {
	ID           uint
	FileID       uint
	Name         string
	NextFileID   uint
	OrderApplied uint
	AppliedAt    *time.Time
	QueryUp      string
	QueryDown    string
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
	Query         *string
}
