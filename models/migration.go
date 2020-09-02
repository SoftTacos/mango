package models

import "time"

// represents a row in the db_version table

type Migration struct {
	ID           uint
	NextID       uint
	OrderApplied uint
	FileID       string
	AppliedAt    *time.Time
}
