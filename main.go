package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

const webDir = "./web"

var Database *sql.DB

const formatDate = "20060102"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	Database, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer Database.Close()

	_, err = os.Stat(dbFile)
	if err != nil {
		createDb(Database)
	} else {
		log.Println("Database already exists")
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Starting server on port: %s\n", port)

	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)
	http.Handle("/api/nextdate", http.HandlerFunc(HandleNextDate))
	http.HandleFunc("/api/task", Check)
	http.HandleFunc("/api/tasks", GetTasks)
	http.HandleFunc("/api/task/done", DoneTask)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
