package main

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	models "github.com/softtacos/mango/models"
)

var migrationFilenameRegex = regexp.MustCompile(`^.*\.sql$`)

func readMigrationFiles(directory string) ([]*models.Migration, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	migrations := []*models.Migration{}
	for _, file := range files {
		filename := file.Name()
		if migrationFilenameRegex.MatchString(filename) {
			migration, err := parseMigrationFile(directory+`\`, filename)
			if err != nil {
				log.Println(err)
				continue
			}
			migrations = append(migrations, &migration)
		}
	}

	return migrations, nil
}

var mangoTagRegex = regexp.MustCompile(`^\s*--mango `)
var whitespaceLineRegex = regexp.MustCompile(`^\s*$`)
var ErrInvalidCommand = errors.New("invalid mango tag")
var ErrNoFileID = errors.New("no file ID after next tag")

func parseMigrationFile(directory, filename string) (models.Migration, error) {
	migration := models.NewMigration()

	fileBytes, err := ioutil.ReadFile(directory + filename)
	if err != nil {
		return migration, err
	}

	// splitFilename := strings.SplitN(filename, "_", 2)
	// fileID, _ := strconv.ParseUint(splitFilename[0], 10, 64)
	// migration.FileID = uint(fileID)
	// migration.Name = splitFilename[1]

	migration.Filename = filename

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
	commandString := mangoTagRegex.ReplaceAllString(string(line), "")
	args := strings.Split(commandString, " ")
	if len(args) < 1 {
		return nil
	}
	switch string(args[0]) {
	case "requires":
		if len(args) < 2 {
			return ErrNoFileID
		}
		migration.RequiredFiles = append(migration.RequiredFiles, args[1:len(args)]...)
	case "up":
		migration.Query = true
	case "down":
		migration.Query = false
	default:
		return ErrInvalidCommand
	}

	return nil
}
