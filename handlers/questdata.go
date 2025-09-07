package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func OpenDB() *sql.DB {
	var Db *sql.DB
	psqlstring := os.Getenv("DATABASE_CONN_STRING")
	Db, err := sql.Open("postgres", psqlstring)
	if err != nil {
		log.Fatal(err)
	}

	err = Db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected!")
	return Db
}

func CloseDB(Db *sql.DB) {
	if Db != nil {
		Db.Close()
		fmt.Println("Closed connection")
	}
}
func IsValidSession(sessionID string) bool {
	// TODO: check in your session store
	return sessionID == "valid123" // dummy check
}
func UserDashboard(w http.ResponseWriter, r *http.Request) {
	Db := OpenDB()
	defer CloseDB(Db)
	tmpl, _ := template.ParseFiles("static/dashboard.html")
	tmpl.Execute(w, "hellow")
}
func InsertQuest() {

}
func UpdateQuest() {

}

func DeleteQuest() {

}
