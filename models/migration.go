package models

import "time"

// represents a row in the db_version table

type Migration struct {
	ID uint
	Filename  string
	AppliedAt *time.Time
}

