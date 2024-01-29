package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var tmpl *template.Template

// Gorilla Mux
// sqlite
// html template
// air: https://github.com/cosmtrek/air
func main() {
	// Parse template
	tmpl = template.Must(template.ParseFiles("index.html"))

	r := mux.NewRouter()
	r.HandleFunc("/", Home).Methods("GET")
	log.Println("Listen and serve at :3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func Home(w http.ResponseWriter, r *http.Request) {
	_ = tmpl.Execute(w, nil)
}
