package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Databases struct {
	DBMap map[string]Database
}

func (d *Databases) ImportJSON() {
	databasesConfigFile, err := os.Open("config/databases.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Parsing databases.json...")
	defer databasesConfigFile.Close()

	// Read opened file as a byte array.
	byteValue, _ := ioutil.ReadAll(databasesConfigFile)

	var databases map[string]json.RawMessage
	var databasesArray []DatabaseData

	json.Unmarshal(byteValue, &databases)
	json.Unmarshal(databases["databases"], &databasesArray)

	fmt.Println("----------------")
	fmt.Println("Databases:")
	for i := 0; i < len(databasesArray); i++ {
		var database Database

		fmt.Println(databasesArray[i].Type)
		switch dbType := databasesArray[i].Type; dbType {
		case "mongo":
			database = &MongoDatabase{}
			dbb := database.(*MongoDatabase)
			dbb.Tst = 2222222222
			fmt.Println(dbb.Tst)
		case "postgres":
			database = &PostgresDatabase{}
		default:
			database = nil
		}

		d.DBMap[databasesArray[i].Name] = database

		// fmt.Printf("key[%s] value[%s]\n", k, v)
		fmt.Println(database)
	}
	fmt.Println(d.DBMap)
	fmt.Println("----------------")
}
