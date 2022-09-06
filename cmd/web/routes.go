package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice" //use this package to avoid writing handler chain
)

func (a *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		a.notFound(w)
	})
	fileServer := http.FileServer(http.Dir("./ui/static"))

	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	router.HandlerFunc(http.MethodGet, "/", a.home)
	router.HandlerFunc(http.MethodGet, "/snippet/view/:id", a.snippetView)
	router.HandlerFunc(http.MethodGet, "/snippet/create", a.snippetCreate)
	router.HandlerFunc(http.MethodPost, "/snippet/create", a.snippetCreatePost)

	standard := alice.New(a.recoverPanic, a.logRequest, secureHeaders)
	return standard.Then(router)
}
