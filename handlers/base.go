package handlers

import (
	"net/http"
)

func New() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/", http.FileServer(http.Dir("static/")))

	// OauthGoogle
	mux.HandleFunc("/auth/google/login", oauthGoogleLogin)
	mux.HandleFunc("/auth/google/callback", oauthGoogleCallback)

	return mux
}
