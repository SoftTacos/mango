package models

import "time"

// represents a row in the db_version table

type Migration struct {
	ID           uint
	FileID       uint
	Name         string
	NextFileID   uint
	OrderApplied uint
	AppliedAt    *time.Time
	QueryUp      string
	QueryDown    string
}
