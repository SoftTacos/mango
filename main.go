package main

import (
	"flag"
	"log"

	gopg "github.com/go-pg/pg/v9"

	models "github.com/softtacos/mango/models"
)

var dbUrl = flag.String("url", "", "url to access the database")

func main() {
	flag.Parse()
	if *dbUrl == "" {
		log.Fatal("please specify a db url")
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
	if result.RowsReturned() == 0{
		createMigrationTable(db)
	} else {
	migrations, err = getDatabaseMigrationData(db)
        	if err != nil {
                	log.Fatal("unable to get database migrations ", err)
        	}
	}
	log.Printf("%+v",result)

	log.Println(migrations)

	
}

func getDatabaseMigrationData(db *gopg.DB) ([]models.Migration, error) {
	migrations := []models.Migration{}
	err := db.Select(&migrations)
	if err != nil {
		return nil, err
	}
	return migrations, nil
}

func createMigrationTable(db *gopg.DB){
	query := `
	CREATE TABLE ()
	`
	_,err:=db.Exec(query)
	if err!=nil{
	log.Println("error creating mango_db_version")
}
}
