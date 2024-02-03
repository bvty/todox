package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var tmpl *template.Template
var db *sql.DB

type Task struct {
	ID        int
	Title     string
	Completed bool
	OOB       string
	IsEditing bool
}

type Stats struct {
	Count          int
	CompletedCount int
}

type Tasks struct {
	Items []Task
	Stats
}

// Gorilla Mux
// sqlite
// html template
// air: https://github.com/cosmtrek/air
func main() {
	// Parse template
	tmpl = template.Must(template.ParseFiles("index.html"))

	// Setup database
	var err error
	db, err = sql.Open("sqlite3", "./todox.db")
	if err != nil {
		log.Fatal("open database failed", err)
	}
	defer db.Close()
	// Initial data
	initialDatabase()

	// Routes setup
	r := mux.NewRouter()
	r.HandleFunc("/", Home).Methods("GET")
	r.HandleFunc("/tasks", Add).Methods("POST")
	r.HandleFunc("/tasks/{id}/toggle", Toggle).Methods("PUT")
	r.HandleFunc("/tasks/{id}", Update).Methods("PUT")
	r.HandleFunc("/tasks/{id}", Del).Methods("DELETE")
	r.HandleFunc("/tasks/{id}", Edit).Methods("GET")
	log.Println("Listen and serve at :3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func Home(w http.ResponseWriter, r *http.Request) {
	ts, err := fetchTasks()
	if err != nil {
		log.Printf("fetch tasks error: %v", err)
		w.Write([]byte("Sever internal error"))
	}
	st, err := statsTasks()
	if err != nil {
		w.Write([]byte("Sever internal error"))
	}
	ts.Stats = st
	_ = tmpl.Execute(w, ts)
}

// Initial database
func initialDatabase() {
	query := `
		create table if not exists tasks (
			id integer primary key autoincrement
		  , title string
		  , completed bool default false
		);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("creating table error: %v", err)
	}
	// Mockup data to tasks
	cnt := 0
	query = `select count(1) from tasks`
	err = db.QueryRow(query).Scan(&cnt)
	if err != nil {
		log.Printf("can not count on table, error: %v", err)
	}
	if cnt == 0 {
		query = `
			insert into tasks (title) values ('Create tasks list');
			insert into tasks (title) values ('Implement CRUD');
			insert into tasks (title) values ('Support HTMX');
			insert into tasks (title) values ('Add Bootstrap');
		`
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("mockup data failed, error: %v", err)
		}
	}
}

// Get tasks list
func fetchTasks() (Tasks, error) {
	var ts Tasks
	query := `select id, title, completed from tasks`
	rows, err := db.Query(query)
	if err != nil {
		return ts, err
	}
	defer rows.Close()
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Title, &t.Completed)
		if err != nil {
			return ts, err
		}
		ts.Items = append(ts.Items, t)
	}
	return ts, nil
}

func statsTasks() (Stats, error) {
	var st Stats
	query := `select coalesce(count(1), 0) as num_of_tasks
				   , coalesce(sum(case when completed then 1 else 0 end), 0) as completed_tasks
				from tasks
	`
	err := db.QueryRow(query).Scan(&st.Count, &st.CompletedCount)
	if err != nil {
		log.Printf("stats tasks error: %v", err)
		return st, err
	}
	return st, nil
}

// Add task
func addTask(task Task) (Task, error) {
	query := `insert into tasks (title) values (?) returning id`
	t := task
	err := db.QueryRow(query, t.Title).Scan(&t.ID)
	if err != nil {
		log.Printf("insert task error: %v", err)
		return t, err
	}
	return t, nil
}

// Add task handler
func Add(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		log.Printf("Title can not be nil")
		return
	}
	t := Task{Title: title, Completed: false}
	log.Printf("inserting task: %s", t.Title)
	task, err := addTask(t)
	if err != nil {
		log.Printf("add task error: %v", err)
		return
	}
	task.OOB = "beforeend:#tasks-list"
	st, err := statsTasks()
	if err != nil {
		log.Printf("stats tasks error: %v", err)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "add-form-block", nil)
	_ = tmpl.ExecuteTemplate(w, "tasks-stats-block", st)
	_ = tmpl.ExecuteTemplate(w, "task-item-block", task)
}

// Update tasks
func updateTask(task Task) error {
	query := `update tasks set title = ?, completed = ? where id = ?`
	_, err := db.Exec(query, task.Title, task.Completed, task.ID)
	log.Printf("updating: %d - %s - %t", task.ID, task.Title, task.Completed)
	if err != nil {
		log.Printf("updating task error: %v", err)
		return err
	}
	return nil
}

func getTask(id int) (Task, error) {
	query := `select id, title, completed from tasks where id = ?`
	var task Task
	err := db.QueryRow(query, id).Scan(&task.ID, &task.Title, &task.Completed)
	if err != nil {
		log.Printf("get task error: %v", err)
		return task, err
	}
	return task, nil
}

func Toggle(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	taskId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("id not found: %v", err)
		return
	}
	t, err := getTask(taskId)
	if err != nil {
		log.Printf("task not found: %v", err)
		return
	}
	t.Completed = !t.Completed
	err = updateTask(t)
	if err != nil {
		log.Printf("updating task error: %v", err)
		return
	}
	t.OOB = "true"
	ts, err := statsTasks()
	if err != nil {
		log.Printf("stats task error: %v", err)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "tasks-stats-block", ts)
	_ = tmpl.ExecuteTemplate(w, "task-item-block", t)
}

// Delete task
func deleteTask(id int) error {
	query := `delete from tasks where id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func Del(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	taskId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("id not found: %v", err)
		return
	}
	t, err := getTask(taskId)
	if err != nil {
		log.Printf("task not found: %v", err)
		return
	}
	err = deleteTask(t.ID)
	if err != nil {
		log.Printf("delete task error: %v", err)
		return
	}
	t.OOB = "delete"
	ts, err := statsTasks()
	if err != nil {
		log.Printf("stats task error: %v", err)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "tasks-stats-block", ts)
	_ = tmpl.ExecuteTemplate(w, "task-item-block", t)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	taskId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("id not found: %v", err)
		return
	}
	t, err := getTask(taskId)
	if err != nil {
		log.Printf("task not found: %v", err)
		return
	}
	t.IsEditing = true
	t.OOB = "true"
	_ = tmpl.ExecuteTemplate(w, "task-item-block", t)
}

func Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	taskId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("id not found: %v", err)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		log.Printf("title can not null")
		return
	}

	t, err := getTask(taskId)
	if err != nil {
		log.Printf("task not found: %v", err)
		return
	}
	t.Title = title

	err = updateTask(t)
	if err != nil {
		log.Printf("updating task error: %v", err)
		return
	}
	t.OOB = "true"

	_ = tmpl.ExecuteTemplate(w, "task-item-block", t)
}
