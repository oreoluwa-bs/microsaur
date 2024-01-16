package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	r.Post("/{databaseId}/query", sendDatabaseQueryById)
	r.Post("/{databaseId}/mutation", sendDatabaseMutationById)
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
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}

func sendDatabaseMutationById(w http.ResponseWriter, r *http.Request) {

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
		return
	}

	db := ConnectToDB(database.Name)

	defer db.Close()

	var body CommandDatabase

	dec := json.NewDecoder(r.Body)
	dec.Decode(&body)

	var interfaceValues []interface{}

	interfaceValues = append(interfaceValues, body.Params...)

	statement, err := db.Prepare(body.SQL)
	// result, err := db.Exec(body.SQL, interfaceValues...)

	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := statement.Exec(interfaceValues...)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.Encode(result)
}

func sendDatabaseQueryById(w http.ResponseWriter, r *http.Request) {

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
		return
	}

	db := ConnectToDB(database.Name)

	defer db.Close()

	var body CommandDatabase

	dec := json.NewDecoder(r.Body)
	dec.Decode(&body)

	var interfaceValues []interface{}
	interfaceValues = append(interfaceValues, body.Params...)

	rows, err := db.Query(body.SQL, interfaceValues...)
	if err != nil {
		fmt.Println("Error executing query:", err)
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fmt.Println("SQL Query:", body.SQL)
	// fmt.Println("SQL Params:", body.Params)

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		w.Write([]byte("Cannot get column names"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create a slice to hold the values for a single row
	values := make([]interface{}, len(columns))

	// Create a slice to hold pointers to the values
	valuePointers := make([]interface{}, len(columns))
	for i := range values {
		valuePointers[i] = &values[i]
	}

	// fmt.Println(values...)
	// fmt.Println(valuePointers...)

	data := make([]map[string]interface{}, 0)

	for rows.Next() {
		err := rows.Scan(valuePointers...)
		if err != nil {
			w.Write([]byte("Cannot get databases"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		row := make(map[string]interface{})
		for i, columnName := range columns {
			row[columnName] = values[i]
		}

		data = append(data, row)
	}

	enc := json.NewEncoder(w)
	enc.Encode(data)
	w.WriteHeader(http.StatusOK)
}
