package db

import (
	gopg "github.com/go-pg/pg/v9"

	models "github.com/softtacos/mango/models"
)

func InsertMigration(db *gopg.DB, migration *models.MigrationDB) error {
	return db.Insert(migration)
}
