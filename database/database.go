package database

import (
	"database/sql"
	"errors"
)

type DatabaseStore struct {
	DB *sql.DB
}

func NewDatabaseStore(db *sql.DB) *DatabaseStore {
	return &DatabaseStore{DB: db}
}

type Database struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateDatabaseDTO struct {
	Name string `json:"name"`
}

func (s *DatabaseStore) Create(body CreateDatabaseDTO) (Database, error) {
	var database Database
	statement, err := s.DB.Prepare(`INSERT INTO databases Values(NULL,?)`)
	if err != nil {
		return database, err
	}

	// Validate
	if body.Name == "main" {
		return database, errors.New(body.Name + " is a reserved word")
	}

	res, err := statement.Exec(body.Name)
	if err != nil {
		return database, err
	}

	//
	newDb := ConnectToDB(body.Name) // creates db
	defer newDb.Close()

	id, err := res.LastInsertId()
	if err != nil {
		return database, err
	}

	database, err = s.GetById(id)
	if err != nil {
		return database, err
	}

	return database, err
}

func (s *DatabaseStore) GetAll() ([]Database, error) {
	rows, err := s.DB.Query(`SELECT * FROM databases`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	data := []Database{}
	for rows.Next() {
		i := Database{}
		err = rows.Scan(&i.ID, &i.Name)
		if err != nil {
			return nil, err
		}
		data = append(data, i)
	}

	return data, nil
}

func (s *DatabaseStore) GetById(id int64) (Database, error) {
	var database Database

	row := s.DB.QueryRow(`SELECT * FROM databases WHERE id = ?`, id)

	if err := row.Scan(&database.ID, &database.Name); err != nil {
		return database, err
	}

	return database, nil
}

type CommandDatabase struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}

func (s *DatabaseStore) SendMutation(id int64, cmd CommandDatabase) (sql.Result, error) {

	var database Database

	row := s.DB.QueryRow("SELECT * FROM databases WHERE id = ?", id)

	if err := row.Scan(&database.ID, &database.Name); err != nil {

		if err == sql.ErrNoRows {
			// Handle case when no rows are returned
			return nil, errors.New("no rows found for the specified ID")
		}

		return nil, err
	}

	db := ConnectToDB(database.Name)

	defer db.Close()

	var interfaceValues []interface{}
	interfaceValues = append(interfaceValues, cmd.Params...)
	// result, err := db.Exec(cmd.SQL, interfaceValues...)
	statement, err := db.Prepare(cmd.SQL)
	if err != nil {
		return nil, err
	}

	result, err := statement.Exec(interfaceValues...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DatabaseStore) SendQuery(id int64, cmd CommandDatabase) ([]map[string]interface{}, error) {
	var database Database

	row := s.DB.QueryRow("SELECT * FROM databases WHERE id = ?", id)

	if err := row.Scan(&database.ID, &database.Name); err != nil {
		return nil, err
	}

	db := ConnectToDB(database.Name)

	defer db.Close()

	var interfaceValues []interface{}
	interfaceValues = append(interfaceValues, cmd.Params...)

	rows, err := db.Query(cmd.SQL, interfaceValues...)
	if err != nil {
		return nil, err
	}

	// fmt.Println("SQL Query:", cmd.SQL)
	// fmt.Println("SQL Params:", cmd.Params)

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
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
			return nil, err
		}

		row := make(map[string]interface{})
		for i, columnName := range columns {
			row[columnName] = values[i]
		}

		data = append(data, row)
	}

	return data, nil
}
