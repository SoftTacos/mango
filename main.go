package main

import (
	"flag"
	"fmt"
	"log"

	gopg "github.com/go-pg/pg/v9"

	queries "github.com/softtacos/mango/db"
	models "github.com/softtacos/mango/models"
)

var dbUrl = flag.String("db", "", "url to access the database")
var migrationDir = flag.String("dir", "", "directory that the migration files are in")
var migrationsRequested = flag.String("mig", "", "names of migrations you would like to apply or remove")
var command = flag.String("cmd", "apply", "apply/remove")

type Command struct {
	Direction     int
	TargetVersion string
}

// idea is to not worry about the actual linear version of the database, but rather to make sure migrations that need to have an order, are applied/removed in that specific order.
// conceptually similar to branching in github except without a branch name. "branches" are defined by the next file ID supplied with the next tag

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

	cmd := Command{}

	err = Migrate(cmd, db, *migrationDir)

}

func Migrate(cmd Command, db *gopg.DB, migrationDir string) error {
	dbMigrations := []models.Migration{}
	//newTable := false
	result, err := db.Exec(`SELECT * FROM pg_tables WHERE tablename = 'mango_db_versions'`)
	if err != nil {
		log.Fatal("unable to get database migrations ", err)
	}

	if result.RowsReturned() == 0 {
		log.Println("no mango_db_versions table found, creating")
		err = queries.CreateMigrationTable(db)
		if err != nil {
			log.Println("error creating mango_db_versions", err)
		}
	} else {
		log.Println("table exists, getting existing migration entries")
		dbMigrations, err = queries.GetDatabaseMigrationData(db)
		if err != nil {
			log.Fatal("unable to get database migrations ", err)
		}
	}

	dbMigMap := map[uint]models.Migration{}
	for _, mig := range dbMigrations {
		dbMigMap[mig.FileID] = mig
	}

	fmt.Printf("DB: %+v", dbMigrations)

	migrationFiles, err := readMigrationFiles(migrationDir)
	if err != nil {
		log.Fatal("unable to get migration files ", err)
	}

	newVersions := []models.Migration{}
	// check if each migration file is in DB
	for _, dbFile := range migrationFiles {
		// if dbFile is not in the database, add to new versions
		if _, ok := dbMigMap[dbFile.FileID]; !ok {
			// TODO: check if we even want to apply it via the commands
			newVersions = append(newVersions, dbFile)
		}
	}
	// if file has migration in DB, check if DB has next that matches, if not update

	// if file doesn't have migration in DB, apply it
	// // TODO: up/downgrade to particular mango version

	// if migration doesn't have file...?

	return nil
}
