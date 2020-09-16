package db

import (
	gopg "github.com/go-pg/pg/v9"

	models "github.com/softtacos/mango/models"
)

func GetDatabaseMigrationData(db *gopg.DB) ([]*models.Migration, error) {
	migrations := []*models.Migration{}
	query := `
		SELECT
			*
		FROM
		mango_db_versions`
	_, err := db.Query(&migrations, query)
	if err != nil {
		return nil, err
	}
	return migrations, nil
}

func CreateMigrationTable(db *gopg.DB) error {
	query := `
	CREATE TABLE mango_db_versions(
		id SERIAL PRIMARY KEY,
		filename TEXT,
		required_files TEXT[],
		order_applied INTEGER,
		query_up TEXT,
		query_down TEXT,
		applied boolean,
		last_applied_at TIMESTAMP WITH TIME ZONE
	)`
	_, err := db.Exec(query)
	return err
}
