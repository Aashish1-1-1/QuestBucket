package handlers

import (
	"context"
	"net/http"
)

func requireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("oauthstate")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		userid, valid := IsValidSession(cookie.Value)
		if !valid {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		Ctx := context.WithValue(r.Context(), "userId", userid)
		next(w, r.WithContext(Ctx))
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
	mux.HandleFunc("/dashboard", requireLogin(UserDashboard))

	return mux
}
