package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"

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

	log.Printf("DB: %+v", dbMigrations)

	migrationFiles, err := readMigrationFiles(*migrationDir)
	if err != nil {
		log.Fatal("unable to get migration files ", err)
	}

	log.Printf("files: %+v", migrationFiles)

	// check next migration ID validities
}

var migrationFilenameRegex = regexp.MustCompile(`^[0-9]{1,15}_.*\.sql$`)

func readMigrationFiles(directory string) ([]models.Migration, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		filename := file.Name()
		if migrationFilenameRegex.MatchString(filename) {
			// migFiles = append(migFiles, file)
			migration, err := parseMigrationFile(directory + filename)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(migration)
		}
	}

	return nil, nil
}

var mangoTagRegex = regexp.MustCompile(`^\s*--mango .*`)
var ErrInvalidCommand = errors.New("invalid mango tag")
var ErrNoFileID = errors.New("no file ID after next tag")

func parseMigrationFile(filename string) (*models.Migration, error) {
	log.Println("parsing file: ", filename)
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	migration := &models.Migration{}

	reader := bytes.NewReader(fileBytes)
	scanner := bufio.NewReader(reader)
	for line, _, err := scanner.ReadLine(); err == nil; line, _, err = scanner.ReadLine() {
		log.Println(string(line))
		if mangoTagRegex.Match(line) {
			parseTag(line, migration)
		}

	}
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}

func parseTag(line []byte, migration *models.Migration) error {
	log.Println("comment line: ", string(line))
	commandBytes := mangoTagRegex.ReplaceAll(line, []byte{})
	args := bytes.Split(commandBytes, []byte{' '})
	if len(args) < 1 {
		return nil // todo: better handling of useless line?
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
	default:
		return ErrInvalidCommand
	}

	return nil
}
