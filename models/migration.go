package models

import "time"

// represents a row in the db_version table

type MigrationDB struct {
	ID            uint
	Filename      string
	RequiredFiles []string `pg:"required_files,array"` // Filenames of the migrations that must be run before this migration is applied
	OrderApplied  uint     //
	Applied       bool     //
	LastAppliedAt *time.Time
	QueryUp       []byte // query to upgrade version
	QueryDown     []byte // query to downgrade version
}

func NewMigration() Migration {
	migDB := &MigrationDB{
		RequiredFiles: []string{},
	}
	return Migration{
		MigrationDB:  migDB,
		Query:        true,
		Dependencies: []*Migration{},
	}
}

type Migration struct {
	*MigrationDB
	// NextMigration *Migration
	Query        bool
	Dependencies []*Migration
}
