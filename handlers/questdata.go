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

type Quest struct {
	Queststitle      sql.NullString
	Questdescription sql.NullString
	Questtag         sql.NullString
}
type Dashboard struct {
	Username string
	Pfp_url  string
	Quests   []Quest
}

func UserDashboard(w http.ResponseWriter, r *http.Request) {
	Db := OpenDB()
	defer CloseDB(Db)
	userId := r.Context().Value("userId").(string)

	var datas Dashboard
	err := Db.QueryRow(`SELECT username, pfp_url FROM public.users WHERE id=$1`, userId).
		Scan(&datas.Username, &datas.Pfp_url)
	if err != nil {
		fmt.Println("Error baby", err)
	}

	rows, err := Db.Query(`SELECT title, description, tags FROM public.quests WHERE user_id=$1`, userId)
	defer rows.Close()
	if err != nil {
		fmt.Println("Error", err)
	}
	for rows.Next() {
		var quest Quest
		if err := rows.Scan(&quest.Queststitle, &quest.Questdescription, &quest.Questtag); err != nil {
			fmt.Println("Error scanning row:", err)
			return
		}
		datas.Quests = append(datas.Quests, quest)
	}
	tmpl, _ := template.ParseFiles("static/dashboard.html")
	err = tmpl.Execute(w, datas)
	if err != nil {
		fmt.Println("Error during tmpl exec", err)
	}
}
func AddQuest() {

}
func UpdateQuest() {

}

func DeleteQuest() {

}
