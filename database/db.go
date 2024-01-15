package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func ConnectToDB(name string) *sql.DB {
	filePath := "data/" + name + ".db"
	os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	db, err := sql.Open("sqlite3", filePath)

	fmt.Println(name)

	if err != nil {
		log.Fatal(err)
	}
	return db
}
