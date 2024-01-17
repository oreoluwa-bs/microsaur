package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oreoluwa-bs/microsaur/database"
)

func main() {
	r := newStore()
	log.Fatal(http.ListenAndServe(":8000", r))
}

func newStore() chi.Router {

	db := database.ConnectToDB("main")

	// Run migrations
	const create string = `
  CREATE TABLE IF NOT EXISTS databases (
  id INTEGER NOT NULL PRIMARY KEY,
  name VARCHAR
  );`

	if _, err := db.Exec(create); err != nil {
		log.Panic(err)
	}

	// defer db.Close()

	ds := database.NewDatabaseStore(db)

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	// r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	templ := template.Must(template.ParseFiles("templates/partials/base.html", "templates/index.html"))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		type data struct {
			Databases []database.Database
		}
		w.Header().Add("Content-type", "text/html")

		Databases, err := ds.GetAll()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		err = templ.Execute(w, data{Databases})
		if err != nil {
			log.Panic(err)
		}
	})

	r.Route("/database", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var body database.CreateDatabaseDTO

			dec := json.NewDecoder(r.Body)
			dec.Decode(&body)

			data, err := ds.Create(body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusCreated)
			enc := json.NewEncoder(w)
			enc.Encode(&data)
		})
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {

			data, err := ds.GetAll()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			enc.Encode(&data)
		})

		// // Regexp url parameters:
		r.Get("/{databaseId}", func(w http.ResponseWriter, r *http.Request) {
			databaseId := Atoi(chi.URLParam(r, "databaseId"))

			data, err := ds.GetById(databaseId)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			enc.Encode(&data)
		})

		r.Post("/{databaseId}/query", func(w http.ResponseWriter, r *http.Request) {

			databaseId := Atoi(chi.URLParam(r, "databaseId"))

			var body database.CommandDatabase

			dec := json.NewDecoder(r.Body)
			dec.Decode(&body)

			data, err := ds.SendQuery(databaseId, body)
			if err != nil {
				fmt.Println(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			enc.Encode(&data)
		})
		r.Post("/{databaseId:[0-9]+}/mutation", func(w http.ResponseWriter, r *http.Request) {
			databaseId := Atoi(chi.URLParam(r, "databaseId"))

			var body database.CommandDatabase

			dec := json.NewDecoder(r.Body)
			dec.Decode(&body)

			data, err := ds.SendMutation(databaseId, body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			enc.Encode(&data)
		})
	})

	return r
}

func Atoi(s string) int64 {
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Panic(err)
	}

	return int64(n)
}
