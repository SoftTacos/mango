package main

import (
	"flag"
	"io/ioutil"
	"log"

	gopg "github.com/go-pg/pg/v9"

	models "github.com/softtacos/mango/models"
)

var dbUrl = flag.String("db", "", "url to access the database")
var migrationDir = flag.String("dir", "", "directory that the migration files are in")

func main() {
	flag.Parse()
	if *dbUrl == "" {
		log.Fatal("please specify a db url")
	}
	if *migrationDir == "" {
		log.Fatal("please specify a migration directory")
	}

	options, err := gopg.ParseURL(*dbUrl)
	if err != nil {
		log.Fatal("unable to connect to the database ", err)
	}

	db := gopg.Connect(options)

	migrations := []models.Migration{}
	//newTable := false
	result, err := db.Exec(`SELECT * FROM pg_tables WHERE tablename = 'mango_db_version'`)
	if err != nil {
		log.Fatal("unable to get database migrations ", err)
	}

	if result.RowsReturned() == 0 {
		err = createMigrationTable(db)
		if err != nil {
			log.Println("error creating mango_db_version", err)
		}
	} else {
		log.Println("table exists, getting existing migration entries")
		migrations, err = getDatabaseMigrationData(db)
		if err != nil {
			log.Fatal("unable to get database migrations ", err)
		}
	}

	log.Printf("%+v", migrations)

	readMigrationFiles(*migrationDir)
}

func getDatabaseMigrationData(db *gopg.DB) ([]models.Migration, error) {
	migrations := []models.Migration{}
	query := `
		SELECT
			*
		FROM
		mango_db_version`
	_, err := db.Query(&migrations, query)
	if err != nil {
		return nil, err
	}
	return migrations, nil
}

func createMigrationTable(db *gopg.DB) error {
	query := `
	CREATE TABLE mango_db_version(
		id SERIAL PRIMARY KEY,
		file_id VARCHAR(255),
		next_id INTEGER,
		order_applied INTEGER,
		applied_at TIMESTAMP WITH TIME ZONE
	)`
	_, err := db.Exec(query)
	return err
}

func readMigrationFiles(directory string) ([]models.Migration, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	log.Println(files)

	return nil, nil
}
