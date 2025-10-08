package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/lib/pq"
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
func InitializeDb() {
	db := OpenDB()
	defer CloseDB(db)
	initializeschema := `
			CREATE TABLE IF NOT EXISTS public.users (
			  id VARCHAR(255) NOT NULL,
			  username VARCHAR(100),
			  email VARCHAR(255),
			  pfp_url VARCHAR(255),
			  CONSTRAINT users_pkey PRIMARY KEY (id)
			);
			
			CREATE TABLE IF NOT EXISTS public.quests (
			  id VARCHAR(255) NOT NULL DEFAULT gen_random_uuid()::text,
			  user_id VARCHAR(255),
			  title VARCHAR(255),
			  description text,
			  tags text[],
			  notes text,
			  CONSTRAINT quests_pkey PRIMARY KEY (id),
			  CONSTRAINT quests_user_id_fkey FOREIGN KEY (user_id)
			    REFERENCES public.users(id)
			    ON DELETE CASCADE
			);
	`
	if _, err := db.Exec(initializeschema); err != nil {
		log.Fatalf("failed to init tables: %v", err)
	}
	fmt.Println("Initialized Successfully")
}

type Quest struct {
	Questsid         sql.NullString
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

	rows, err := Db.Query(`SELECT id,title, description, tags FROM public.quests WHERE user_id=$1`, userId)
	defer rows.Close()
	if err != nil {
		fmt.Println("Error", err)
	}
	for rows.Next() {
		var quest Quest
		if err := rows.Scan(&quest.Questsid, &quest.Queststitle, &quest.Questdescription, &quest.Questtag); err != nil {
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
	_, err = Db.Query(`insert into "quests"("user_id","title", "description","tags","notes") values($1, $2, $3,$4,$5)`, userId, data.Title, data.Description, pq.Array(data.Tags), data.Note)

	if err != nil {
		fmt.Println("Error occured", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func DeleteQuest(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 {
		Db := OpenDB()
		defer CloseDB(Db)
		id := parts[2]
		userId, ok := r.Context().Value("userId").(string)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		query := `DELETE FROM "quests" WHERE "id"=$1 AND "user_id"=$2`
		res, err := Db.Exec(query, id, userId)
		if err != nil {
			http.Error(w, "Sql Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		affected, err := res.RowsAffected()
		if err != nil {
			http.Error(w, "Sql Error2: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if affected == 0 {
			fmt.Println("No rows deleted")
		} else {
			fmt.Printf("%d row(s) deleted\n", affected)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
		return
	} else {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}
}

func GetNotes(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 {
		id := parts[2]
		Db := OpenDB()
		defer CloseDB(Db)
		var markdown string
		err := Db.QueryRow(`Select notes from "quests" where id=$1`, id).Scan(&markdown)
		if err != nil {
			fmt.Println("Error occured", err)
		}
		html := mdToHTML([]byte(markdown))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		w.Write(html)
	} else {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}
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
func EditPost(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 {
		id := parts[3]
		Db := OpenDB()
		defer CloseDB(Db)
		if r.Method == http.MethodPost {
			userId, ok := r.Context().Value("userId").(string)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if err := r.ParseForm(); err != nil {
				http.Error(w, "parse error: "+err.Error(), http.StatusBadRequest)
				return
			}

			markdown := r.FormValue("content")
			query := `UPDATE "quests" SET "notes"=$1 WHERE "id"=$2 AND "user_id"=$3`
			res, err := Db.Exec(query, markdown, id, userId)
			if err != nil {
				http.Error(w, "Sql Error1: "+err.Error(), http.StatusInternalServerError)
				return
			}

			affected, err := res.RowsAffected()
			if err != nil {
				http.Error(w, "Sql Error2: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if affected == 0 {
				fmt.Println("No rows updated")
			} else {
				fmt.Printf("%d row(s) updated\n", affected)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
			})
			return
		} else if r.Method == http.MethodGet {
			var markdown string
			err := Db.QueryRow(`Select notes from "quests" where id=$1`, id).Scan(&markdown)
			if err != nil {
				fmt.Println("Error occured", err)
			}
			tmpl := template.Must(template.New("edior").Parse(page))
			datas := struct {
				Content string
				Id      string
			}{
				Content: markdown,
				Id:      id,
			}
			err = tmpl.Execute(w, datas)
			if err != nil {
				fmt.Println("Error during tmpl exec", err)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}
}

func Profile(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 3 {
		username := parts[2]
		Db := OpenDB()
		defer CloseDB(Db)
		var datas Dashboard
		datas.Username = username

		var userId string
		err := Db.QueryRow(`SELECT id ,pfp_url FROM public.users WHERE username=$1`, username).
			Scan(&userId, &datas.Pfp_url)
		if err != nil {
			fmt.Println("Error baby", err)
		}

		rows, err := Db.Query(`SELECT id,title, description, tags FROM public.quests WHERE user_id=$1`, userId)
		defer rows.Close()
		if err != nil {
			fmt.Println("Error", err)
		}
		for rows.Next() {
			var quest Quest
			if err := rows.Scan(&quest.Questsid, &quest.Queststitle, &quest.Questdescription, &quest.Questtag); err != nil {
				fmt.Println("Error scanning row:", err)
				return
			}
			datas.Quests = append(datas.Quests, quest)
		}
		tmpl := template.Must(template.New("edior").Parse(profile))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err = tmpl.Execute(w, datas)
		if err != nil {
			fmt.Println("Error during tmpl exec", err)
		}
	} else {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}
}
