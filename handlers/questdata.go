package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"os"
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
	QuestNote        sql.NullString
	Questtag         pq.StringArray
}

type Questjson struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Note        string   `json:"note"`
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
func AddQuest(w http.ResponseWriter, r *http.Request) {
	Db := OpenDB()
	defer CloseDB(Db)
	userId := r.Context().Value("userId").(string)
	var data Questjson
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
	}
	_, err = Db.Query(`insert into "quest"("user_id","title", "description","tags","notes",) values($1, $2, $3,$4,$5)`, userId, data.Title, data.Description, data.Tags, data.Note)

	if err != nil {
		fmt.Println("Error occured", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
func UpdateQuest() {

}

func DeleteQuest() {

}

func GetNotes(w http.ResponseWriter, r *http.Request) {

}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
