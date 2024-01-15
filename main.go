package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oreoluwa-bs/microsaur/database"
)

func main() {

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	// r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/database", database.DatabaseRouter)

	log.Fatal(http.ListenAndServe(":8000", r))

}
