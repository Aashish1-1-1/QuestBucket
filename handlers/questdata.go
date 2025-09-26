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
			var page = `
					<!doctype html>
					<html lang="en">
					<head>
					  <meta charset="utf-8" />
					  <meta name="viewport" content="width=device-width, initial-scale=1" />
					  <title>QuestBucket</title>
					
					  <!-- EasyMDE CSS -->
					  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/easymde/dist/easymde.min.css">
					
					  <!-- Tailwind CSS -->
					  <script src="https://cdn.tailwindcss.com"></script>
					
					  <style>
					    body {
					      font-family: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial;
					      background: #f8fafc;
					      padding: 24px;
					    }
					    .container {
					      max-width: 900px;
					      margin: 0 auto;
					    }
					  </style>
					</head>
					<body>
					  <div class="container">
					    <h1 class="text-3xl font-bold text-center mb-6">QuestBucket</h1>
					
					    <!-- Textarea that EasyMDE will enhance -->
					    <textarea id="md" name="md" rows="10">{{ .Content }}</textarea>
					
					    <!-- Save Button -->
					    <div class="mt-6 text-center">
					      <button id="saveBtn" class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded shadow-md transition">
					        Save
					      </button>
					    </div>
					  </div>
					
					  <!-- Toast -->
					  <div id="toast" class="fixed bottom-5 left-1/2 -translate-x-1/2 bg-green-600 text-white px-4 py-2 rounded shadow-lg opacity-0 pointer-events-none transition-opacity duration-500">
					    Success
					  </div>
					
					  <!-- EasyMDE JS -->
					  <script src="https://cdn.jsdelivr.net/npm/easymde/dist/easymde.min.js"></script>
					
					  <script>
					    // Initialize EasyMDE
					    const easyMDE = new EasyMDE({
					      element: document.getElementById('md'),
					      autosave: { enabled: false },
					      spellChecker: false,
					      toolbar: ["bold", "italic", "heading", "|", "quote", "unordered-list", "ordered-list", "|", "link", "image", "|", "preview", "side-by-side", "fullscreen"],
					      autofocus: true,
					    });
					
					    // Toast logic
					    function showToast() {
					      const toast = document.getElementById('toast');
					      toast.classList.remove('opacity-0', 'pointer-events-none');
					      toast.classList.add('opacity-100');
					      setTimeout(() => {
					        toast.classList.add('opacity-0', 'pointer-events-none');
					        toast.classList.remove('opacity-100');
					      }, 5000);
					    }
					
					    // Save button click
					    const saveBtn = document.getElementById('saveBtn');
					    saveBtn.addEventListener('click', () => {
					      const content = easyMDE.value();
					
					      fetch("/edit/post/{{.Id}}", {
					        method: 'POST',
					        headers: {
					          'Content-Type': 'application/x-www-form-urlencoded',
					        },
					        body: new URLSearchParams({ content })
					      })
					      .then(res => {
					        if (!res.ok) throw new Error('Network response was not ok');
					        return res.json();
					      })
					      .then(data => {
					        // Show toast on success
					        showToast();
					      })
					      .catch(err => {
					        console.error('Save failed', err);
					        alert('Save failed: ' + err.message);
					      });
					    });
					
					    // Optional Ctrl+S shortcut
					    window.addEventListener('keydown', function(e) {
					      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
					        e.preventDefault();
					        saveBtn.click();
					      }
					    });
					  </script>
					</body>
					</html>
				`
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
