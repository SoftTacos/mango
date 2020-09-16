package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	gopg "github.com/go-pg/pg/v9"

	queries "github.com/softtacos/mango/db"
	models "github.com/softtacos/mango/models"
)

var dbUrl = flag.String("db", "", "url to access the database")
var migrationDir = flag.String("dir", "", "directory that the migration files are in")
var migrationsRequested = flag.String("mig", "", "names of migrations you would like to apply or remove")
var command = flag.String("cmd", "up", "up or down")
var autoApply = flag.Bool("auto", false, "whether or not to apply pre-requisite migrations that have not been applied for a given migration if they have not been supplied")

type Direction int8

const (
	Up   Direction = iota
	Down Direction = iota
)

type Command struct {
	Command             Direction
	RequestedMigrations map[string]bool
	AutoApply           bool
}

// idea is to not worry about the actual linear version of the database, but rather to make sure migrations that need to have an order, are applied/removed in that specific order.
// conceptually similar to branching in github except without a branch name. "branches" are defined by the next file ID supplied with the next tag

// TODO: support "apply all" migrations for a directory

func main() {
	flag.Parse()

	// check required args
	if *dbUrl == "" {
		log.Fatal("please specify a db url")
	}
	if *migrationDir == "" {
		log.Fatal("please specify a migration directory")
	}
	if *migrationsRequested == "" {
		log.Fatal("please specify migrations to apply")
	}

	// setup DB
	options, err := gopg.ParseURL(*dbUrl)
	if err != nil {
		log.Fatal("unable to connect to the database ", err)
	}

	db := gopg.Connect(options)

	// parse requested migrations
	migrationSlice := strings.Split(*migrationsRequested, " ")
	migrations := map[string]bool{}
	for _, mig := range migrationSlice {
		migrations[mig] = true
	}
	direction := Up
	if *command == "down" {
		direction = Down
	}

	cmd := Command{
		Command:             direction,
		RequestedMigrations: migrations,
		AutoApply:           *autoApply,
	}

	err = Migrate(cmd, db, *migrationDir)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}

var ErrNoMigrations = errors.New("no migrations to run")
var ErrMigrationNotFound = errors.New("one of the migrations requested was not found in the supplied directory")
var ErrMigrationDependencyNotFound = errors.New("a dependency for one of the migrations was not found")

func Migrate(cmd Command, db *gopg.DB, migrationDir string) error {
	dbMigrations := []*models.Migration{}
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

	migrationFiles, err := readMigrationFiles(migrationDir)
	if err != nil {
		log.Fatal("unable to get migration files ", err)
	}
	if len(migrationFiles) == 0 {
		log.Println("no migrations to run in directory:", migrationDir)
		return ErrNoMigrations
	}

	dbMigMap := map[string]*models.Migration{}
	for _, mig := range dbMigrations {
		dbMigMap[mig.Filename] = mig
	}

	newVersions := map[string]*models.Migration{}
	fileMigMap := map[string]*models.Migration{}
	// check if each migration file is in DB
	for _, dbFile := range migrationFiles {
		// if dbFile is not in the database, add to new versions
		if _, ok := dbMigMap[dbFile.Filename]; !ok {
			// TODO: check if we even want to apply it via the commands
			newVersions[dbFile.Filename] = dbFile
		}
		fileMigMap[dbFile.Filename] = dbFile
	}
	// TODO: throw error if database migrationd doesn't have a corresponding file

	// // associate prerequisite migrations
	for _, migration := range fileMigMap {
		for _, id := range migration.RequiredFiles {
			if _, ok := fileMigMap[id]; !ok {
				return ErrMigrationDependencyNotFound
			}
			migration.Dependencies = append(migration.Dependencies, fileMigMap[id])
		}
	}

	for _, mig := range fileMigMap {
		fmt.Println(mig.Filename, mig)
	}

	fmt.Println(cmd.RequestedMigrations)
	fmt.Println(fileMigMap)
	fmt.Println(dbMigMap)
	for requested := range cmd.RequestedMigrations {
		if _, ok := fileMigMap[requested]; !ok {
			return ErrMigrationNotFound
		}

		err := ApplyMigration(requested, fileMigMap, dbMigMap, db)
		if err != nil {
			log.Println(err)
		}
	}
	// if file has migration in DB, check if DB has next that matches, if not update

	// if file doesn't have migration in DB, apply it
	// // TODO: up/downgrade to particular mango version

	// if migration doesn't have file...?

	return nil
}

func ApplyMigration(requested string, fileMigMap, dbMigMap map[string]*models.Migration, db *gopg.DB) error {
	fmt.Println(requested, fileMigMap[requested])
	if dbMigMap[requested].Applied {
		return nil
	}
	for _, dbMig := range fileMigMap[requested].Dependencies {
		err := ApplyMigration(dbMig.Filename, fileMigMap, dbMigMap, db)
		if err != nil {
			return err
		}
	}

	fmt.Println("APPLYING MIGRATION", requested)
	dbMigMap[requested].Applied = true

	return queries.InsertMigration(db, dbMigMap[requested].MigrationDB)
}
