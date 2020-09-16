package models

import (
	"fmt"
	"time"
)

// represents a row in the db_version table

type MigrationDB struct {
	ID            uint
	Filename      string
	RequiredFiles []string `pg:"required_files,array"` // Filenames of the migrations that must be run before this migration is applied
	Applied       bool     //
	LastAppliedAt *time.Time
	QueryUp       []byte // query to upgrade version
	QueryDown     []byte // query to downgrade version

	tableName struct{} `pg:"mango_db_versions"`
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

func (m Migration) String() string {
	return fmt.Sprintf("{%d %s %+v %+v}", m.ID, m.Filename, m.LastAppliedAt, m.Dependencies)
}
