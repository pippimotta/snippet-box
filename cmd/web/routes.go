package main

import (
	"net/http"

	"github.com/justinas/alice" //use this package to avoid writing handler chain
)

func (a *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", a.home)
	mux.HandleFunc("/snippet/view", a.snippetView)
	mux.HandleFunc("/snippet/create", a.snippetCreate)

	standard := alice.New(a.recoverPanic, a.logRequest, secureHeaders)
	return standard.Then(mux)
}
