package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	gopg "github.com/go-pg/pg/v9"

	queries "github.com/softtacos/mango/db"
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

	err = Migrate(db, *migrationDir)

}

func Migrate(db *gopg.DB, migrationDir string) error {
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

	fmt.Printf("DB: %+v", dbMigrations)

	migrationFiles, err := readMigrationFiles(migrationDir)
	if err != nil {
		log.Fatal("unable to get migration files ", err)
	}
	for _, mig := range migrationFiles {
		fmt.Printf("UP QUERY: %+v\n", mig.MigrationDB)
		// fmt.Printf("DOWN QUERY: %+v\n", string(mig.MigrationDB.QueryDown))
	}
	// check next migration ID validities

	// goto each file, check if migration is in DB

	//

	return nil
}

var migrationFilenameRegex = regexp.MustCompile(`^[0-9]{1,20}_.*\.sql$`)

func readMigrationFiles(directory string) ([]models.Migration, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	migrations := []models.Migration{}
	for _, file := range files {
		filename := file.Name()
		if migrationFilenameRegex.MatchString(filename) {
			migration, err := parseMigrationFile(directory, filename)
			if err != nil {
				log.Println(err)
				continue
			}
			migrations = append(migrations, migration)
		}
	}

	return migrations, nil
}

var mangoTagRegex = regexp.MustCompile(`^\s*--mango `)
var whitespaceLineRegex = regexp.MustCompile(`^\s*$`)
var ErrInvalidCommand = errors.New("invalid mango tag")
var ErrNoFileID = errors.New("no file ID after next tag")

func parseMigrationFile(directory, filename string) (models.Migration, error) {
	log.Println("parsing file: ", filename)
	migration := models.NewMigration()

	fileBytes, err := ioutil.ReadFile(directory + filename)
	if err != nil {
		return migration, err
	}

	splitFilename := strings.SplitN(filename, "_", 2)
	fileID, _ := strconv.ParseUint(splitFilename[0], 10, 64)
	migration.FileID = uint(fileID)
	migration.Name = splitFilename[1]

	reader := bytes.NewReader(fileBytes)
	scanner := bufio.NewReader(reader)
	for line, _, err := scanner.ReadLine(); err == nil; line, _, err = scanner.ReadLine() {
		if whitespaceLineRegex.Match(line) {
			continue
		}
		// fmt.Println(string(line))
		if mangoTagRegex.Match(line) {
			err = parseTag(line, &migration)
			if err != nil {
				return migration, err
			}
			continue
		}

		if migration.Query { // up query
			migration.QueryUp = append(migration.QueryUp, append(line, []byte("\n")...)...)
		} else { // down query
			migration.QueryDown = append(migration.QueryDown, append(line, []byte("\n")...)...)
		}
	}

	return migration, nil
}

func parseTag(line []byte, migration *models.Migration) error {
	log.Println("comment line: ", string(line))
	commandBytes := mangoTagRegex.ReplaceAll(line, []byte{})
	log.Println("LINE: ", string(commandBytes))
	args := bytes.Split(commandBytes, []byte(" "))
	if len(args) < 1 {
		return nil
	}
	switch string(args[0]) {
	case "next":
		if len(args) < 2 {
			return ErrNoFileID
		}
		nfid64, err := strconv.ParseUint(string(args[1]), 10, 64)
		migration.NextFileID = uint(nfid64)
		if err != nil {
			return err
		}
	case "up":
		log.Println("UP")
		// migration.Query = &migration.MigrationDB.QueryUp
		migration.Query = true
	case "down":
		log.Println("DOWN")
		// migration.Query = &migration.MigrationDB.QueryDown
		migration.Query = false
	default:
		return ErrInvalidCommand
	}

	return nil
}
