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

const htmlTemplate = `
Threads
<form id="create-db-form" method="POST">
<label>
name
<input name="name" id='name' placeholder= "name of database"></input>
</label>

<button>Submit</button>
</form>

<form id="sql-form" method="POST">
<label>
Database
<select name="id" id="id" required>
</select>
</label>

<label>
Type
<select name="type" value="query" id="type" required>
<option value="query">Query</option>
<option value="mutation">Mutation</option>
</select>
</label>

<label>
Query
<textarea name="sql" id='sql' required></textarea>
</label>

<label>
Params
<input name="params" id='params' placeholder= "1,2,3"></input>
</label>

<button>Submit</button>
</form>

<div id='answer'></div>

<script>
  document.addEventListener("DOMContentLoaded", () => {
    var sqlForm = document.querySelector("form#sql-form");
    createDb = document.querySelector("form#create-db-form");
    const answr = document.querySelector("#answer");
    const idSelector = sqlForm?.querySelector("select#id");

    // console.log(idSelector)

    sqlForm?.addEventListener("submit", async (e) => {
      e.preventDefault();
      if (answr) {
        answr.innerHTML = "";
      }

      const formData = new FormData(e.target);
      const formObject = Object.fromEntries(formData);

      //   console.log(e.target.action);
      //   console.log(formData.get("sql"));
      //   console.log(formObject);
      formObject.params = formObject.params.split(",");
      const action = "/database/" + formObject.id + "/" + formObject.type;

      const response = await fetch(action, {
        method: "POST",
        body: JSON.stringify(formObject),
      });

      if (!response.ok) {
        const text = await response.text();
        alert(text);
        return;
      }

      const result = await response.json();
      if (answr) {
        answr.innerHTML = JSON.stringify(result, undefined, 2);
      }
    });

    createDb?.addEventListener("submit", async (e) => {
      e.preventDefault();

      const formData = new FormData(e.target);
      const formObject = Object.fromEntries(formData);
      const action = "/database";

      const response = await fetch(action, {
        method: "POST",
        body: JSON.stringify(formObject),
      });

      if (!response.ok) {
        const text = await response.text();
        alert(text);
        return;
      }

      const result = await response.json();
    //   alert(response.statusText + ": " + JSON.stringify(result));

      const newOption = document.createElement("option");
      newOption.innerText = result.name;
      newOption.value = result.id;
      idSelector.appendChild(newOption);
    });

    fetch("/database")
      .then(async (resp) => {
        if (!resp.ok) {
          const text = await resp.text();
          throw new Error(text);
        }
        return resp.json();
      })
      .then((result) => {
        result.forEach((r) => {
          const newOption = document.createElement("option");
          newOption.innerText = r.name;
          newOption.value = r.id;
          idSelector.appendChild(newOption);
        });
      })
      .catch((err) => {
        if (err instanceof Error) {
          alert(err.message);
        }
      });
  });
</script>

`

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

	templ := template.Must(template.New("").Parse(htmlTemplate))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		type data struct {
		}
		w.Header().Add("Content-type", "text/html")
		templ.Execute(w, data{})
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
