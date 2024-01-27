package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var (
	TMPL *template.Template
	DB   *sql.DB
)

type Task struct {
	ID       int
	Title    string
	Comleted bool
}

type Tasks struct {
	Items []Task
}

func main() {
	// Parse template
	TMPL = template.Must(template.ParseFiles("index.html"))

	// Setup database
	var err error
	DB, err = sql.Open("sqlite3", "./todox.db")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()
	err = setupDb()
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/", Home)
	log.Println("Listen and serve at :3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func Home(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{"Name": "cuti"}
	TMPL.Execute(w, data)
}

// Initial data
func setupDb() error {
	// Create a table named: tasks
	query := `
		create table tasks (
			id integer autoincresement primary key
		  , title string
		  , completed bool default false
		)
	`
	_, err := DB.Exec(query)
	if err != nil {
		log.Printf("creating table error: %v", err)
		return err
	}
	i := 0
	query = `select count(1) from tasks`
	err = DB.QueryRow(query).Scan(&i)
	if err != nil {
		log.Printf("counting table error: %v", err)
		return err
	}
	if i == 0 {
		query = `
			insert into tasks (title) values ('Create tasks list');
			insert into tasks (title) values ('Implement CRUD functions');
			insert into tasks (title) values ('Create tasks counter');
			insert into tasks (title) values ('SupportHTMX');
		`
		_, err := DB.Exec(query)
		if err != nil {
			log.Printf("creating table error: %v", err)
			return err
		}
	}
	return nil
}

// Fetch data
func fetchAll() (Tasks, error) {
	var ts Tasks
	query := `select id, title, completed from tasks`
	rows, err := DB.Query(query)
	if err != nil {
		return ts, err
	}
	defer rows.Close()
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Title, &t.Comleted)
		if err != nil {
			log.Printf("fetching data error: %v", err)
			return ts, err
		}
		ts.Items = append(ts.Items, t)
	}
	return ts, nil
}
