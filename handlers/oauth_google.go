package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Userinformation
type Userinfo struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Verified bool   `json:"verified_email"`
	Picture  string `json:"picture"`
}

var googleOauthConfig oauth2.Config
var client redis.Client
var ctx context.Context

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	googleOauthConfig = oauth2.Config{
		RedirectURL:  "http://localhost:8000/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	client = *redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})
	ctx = context.Background()
}

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w)

	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the the state query parameter on your redirect callback.
	*/
	//	googleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	//	googleOauthConfig.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	u := googleOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	oauthState, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// GetOrCreate User in your db.
	// Redirect or response with a token.
	// More code .....
	var userdata Userinfo
	err = json.Unmarshal(data, &userdata)
	if err != nil {
		fmt.Println("Error occured", err)
	}
	//add session
	err = client.Set(ctx, oauthState.Value, userdata.Id, 24*time.Hour).Err()
	if err != nil {
		panic(err)
	}
	//modifying pfp_url cause that url is not downloading first the image
	userdata.Picture = userdata.Picture[:len(userdata.Picture)-4] + "0"
	// Add user to db
	i := strings.Index(userdata.Email, "@")
	Db := OpenDB()
	defer CloseDB(Db)
	var email string
	err = Db.QueryRow("SELECT email FROM users where id=$1", userdata.Id).Scan(&email)
	if err == nil {
		http.Redirect(w, r, "/dashboard", http.StatusPermanentRedirect)
		return
	}
	if err != sql.ErrNoRows {
		return
	}
	_, err = Db.Query(`insert into "users"("id","username", "email","pfp_url") values($1, $2, $3,$4)`, userdata.Id, userdata.Email[:i], userdata.Email, userdata.Picture)

	if err != nil {
		fmt.Println("Error occured", err)
	}
	fmt.Println("Inserted user data")
	http.Redirect(w, r, "/dashboard", http.StatusPermanentRedirect)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Path: "/", Expires: expiration, HttpOnly: true, Secure: false}
	http.SetCookie(w, &cookie)

	return state
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}

func IsValidSession(sessionID string) (string, bool) {
	val, err := client.Get(ctx, sessionID).Result()
	if err != nil {
		return "", false
	}
	return val, true // dummy check
}
func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("oauthstate")
	if err != nil {
		fmt.Println("Error reading cookie")
		return
	}
	cookieset := http.Cookie{Value: "", Path: "/", HttpOnly: true, Secure: false}
	http.SetCookie(w, &cookieset)
	_, err = client.Del(ctx, cookie.Value).Result()
	if err != nil {
		fmt.Println("Error deleting session", err)
		return
	}
	fmt.Println("Logged out")
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
