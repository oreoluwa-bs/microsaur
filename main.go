package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

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
<form method="POST">
<label>
Database Id
<input name="id" id='id'></input>
</label>

<label>
Type
<select name="type" value="query" id="type">
<option value="query">Query</option>
<option value="mutation">Mutation</option>
</select>
</label>

<label>
Query
<textarea name="sql" id='sql' ></textarea>
</label>

<label>
Params
<input name="params" id='params' placeholder= "1,2,3"></input>
</label>

<button>Submit</button>
</form>

<div id='answer'></div>

<script>
var form = document.querySelector("form");
const answr =document.querySelector("#answer");

form?.addEventListener("submit", async (e) => {
  e.preventDefault();
  if(answr){
	  answr.innerHTML = "";
	}

  const formData = new FormData(e.target);
  const formObject = Object.fromEntries(formData);

//   console.log(e.target.action);
//   console.log(formData.get("sql"));
//   console.log(formObject);


formObject.params = formObject.params.split(",");


const action = "/database/"+formObject.id+"/"+formObject.type;

  const response = await fetch(action, {
	method:"POST",
    body: JSON.stringify(formObject),
  });
  
  if (!response.ok) {
	const text = await response.text();
    alert(text);
    return;
  }

  const result = await response.json();
  if(answr){
	  answr.innerHTML = JSON.stringify(result,undefined,2)
	}
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

			if err := ds.Create(body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Created"))
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
			databaseId := chi.URLParam(r, "databaseId")

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

			databaseId := chi.URLParam(r, "databaseId")

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
		r.Post("/{databaseId}/mutation", func(w http.ResponseWriter, r *http.Request) {
			databaseId := chi.URLParam(r, "databaseId")

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
