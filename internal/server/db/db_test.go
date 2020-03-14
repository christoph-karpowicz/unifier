package db

import (
	"log"
	"os"
	"testing"
)

var configs *configArray

func TestYAML(t *testing.T) {
	os.Chdir("../../..")
	dbs := Databases{DBMap: make(map[string]*Database)}
	configs = dbs.ImportYAML()
	dbs.ValidateYAML()
}

func TestMongoCRUD(t *testing.T) {
	var database Database
	for i := 0; i < len(configs.Databases); i++ {
		// fmt.Printf("val: %s\n", configs.Databases[i].Name)
		if dbType := configs.Databases[i].Type; dbType == "mongo" {
			database = &mongoDatabase{cfg: &configs.Databases[i]}
			break
		}
	}

	if database != nil {
		database.Init()

		// Select
		rows := database.Select("Sakila_films", "{\"_id\":{\"$lt\": 3}}")
		log.Println(len(rows))

		// Insert
		row := map[string]interface{}{
			"Title":       "test1",
			"Description": "testdesc",
			"ReleaseYear": 2010,
			"Length":      90,
			"ext_id":      1001,
		}
		_, insertErr := database.Insert("Sakila_films", "film_id", 1, row)
		if insertErr != nil {
			log.Fatalln(insertErr)
		}

		// Update
		_, updateErr := database.Update("Sakila_films", "_id", 6, "Rating", "test")
		if updateErr != nil {
			log.Fatalln(updateErr)
		}
	}
}

func TestPostgresCRUD(t *testing.T) {
	var database Database
	for i := 0; i < len(configs.Databases); i++ {
		// fmt.Printf("val: %s\n", configs.Databases[i].Name)
		if dbType := configs.Databases[i].Type; dbType == "postgres" {
			database = &postgresDatabase{cfg: &configs.Databases[i]}
			break
		}
	}

	if database != nil {
		database.Init()

		// Select
		rows := database.Select("film", "film_id > 10 AND film_id < 22")
		log.Println(len(rows))

		// Insert
		row := map[string]interface{}{
			"title":        "test1",
			"description":  "testdesc",
			"release_year": 2010,
			"length":       90,
			"language_id":  2,
		}
		_, insertErr := database.Insert("film", "_id", 1, row)
		if insertErr != nil {
			log.Fatalln(insertErr)
		}

		// Update
		_, updateErr := database.Update("film", "film_id", 1, "description", "test")
		if updateErr != nil {
			log.Fatalln(updateErr)
		}
	}
}
