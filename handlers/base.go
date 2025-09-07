package handlers

import (
	"net/http"
)

func requireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || !IsValidSession(cookie.Value) {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		next(w, r)
	}
}
func New() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/", http.FileServer(http.Dir("static/")))
	//logo
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))
	// OauthGoogle
	mux.HandleFunc("/auth/google/login", oauthGoogleLogin)
	mux.HandleFunc("/auth/google/callback", oauthGoogleCallback)

	//User's data endpints
	mux.HandleFunc("/dashboard", UserDashboard)

	return mux
}
