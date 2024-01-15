package database

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type DatabaseHandler struct {
}

var mainDb *sql.DB

func DatabaseRouter(r chi.Router) {
	if mainDb == nil {
		mainDb = ConnectToDB("main")

		// Run migrations
		const create string = `
  CREATE TABLE IF NOT EXISTS databases (
  id INTEGER NOT NULL PRIMARY KEY,
  name VARCHAR
  );`

		if _, err := mainDb.Exec(create); err != nil {
			log.Panic(err)
		}

		// defer mainDb.Close()
	}

	r.Post("/", createDatabase)
	r.Get("/", getAllDatabases)

	// Regexp url parameters:
	r.Get("/{databaseId}", getDatabaseById)
	r.Post("/{databaseId}", sendDatabaseCommandById)
}

type Database struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateDatabaseDTO struct {
	Name string `json:"name"`
}

func createDatabase(w http.ResponseWriter, r *http.Request) {

	statement, err := mainDb.Prepare(`INSERT INTO databases Values(NULL,?)`)
	if err != nil {
		w.Write([]byte("Cannot create new database"))
		w.WriteHeader(http.StatusBadRequest)
	}

	var body CreateDatabaseDTO
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&body)

	// Validate
	if body.Name == "main" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(body.Name + " is a reserved word"))
		return
	}

	_, err = statement.Exec(body.Name)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//
	newDb := ConnectToDB(body.Name) // creates db

	defer newDb.Close()

	w.Write([]byte("Created"))
	w.WriteHeader(http.StatusCreated)
}

func getAllDatabases(w http.ResponseWriter, r *http.Request) {

	rows, err := mainDb.Query(`SELECT * FROM databases`)
	if err != nil {
		w.Write([]byte("Cannot get databases"))
		w.WriteHeader(http.StatusBadRequest)
	}

	defer rows.Close()

	data := []Database{}
	for rows.Next() {
		i := Database{}
		err = rows.Scan(&i.ID, &i.Name)
		if err != nil {
			w.Write([]byte("Cannot get databases"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		data = append(data, i)
	}

	enc := json.NewEncoder(w)
	enc.Encode(data)

	w.WriteHeader(http.StatusOK)
}

func getDatabaseById(w http.ResponseWriter, r *http.Request) {

	databaseId := chi.URLParam(r, "databaseId")

	var database Database

	row := mainDb.QueryRow(`SELECT * FROM databases WHERE id = ?`, databaseId)

	if err := row.Scan(&database.ID, &database.Name); err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
	}

	enc := json.NewEncoder(w)
	enc.Encode(database)

	w.WriteHeader(http.StatusOK)
}

type CommandDatabase struct {
	Query  string   `json:"query"`
	Values []string `json:"values"`
}

func sendDatabaseCommandById(w http.ResponseWriter, r *http.Request) {

	databaseId := chi.URLParam(r, "databaseId")

	var database Database

	row := mainDb.QueryRow("SELECT * FROM databases WHERE id = ?", databaseId)

	if err := row.Scan(&database.ID, &database.Name); err != nil {

		if err == sql.ErrNoRows {
			// Handle case when no rows are returned
			w.Write([]byte("No rows found for the specified ID."))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
	}

	db := ConnectToDB(database.Name)

	defer db.Close()

	var body CommandDatabase

	dec := json.NewDecoder(r.Body)
	dec.Decode(&body)

	var interfaceValues []interface{}

	for _, v := range body.Values {
		interfaceValues = append(interfaceValues, v)
	}

	db.Exec(body.Query, interfaceValues...)

	w.WriteHeader(http.StatusOK)
}
