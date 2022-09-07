package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice" //use this package to avoid writing handler chain
)

func (a *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	//create a new middleware chain containing the middleware to control dynamic routes
	dynamic := alice.New(a.sessionManager.LoadAndSave)
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(a.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(a.snippetView))
	router.Handler(http.MethodGet, "/snippet/create", dynamic.ThenFunc(a.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", dynamic.ThenFunc(a.snippetCreatePost))

	standard := alice.New(a.recoverPanic, a.logRequest, secureHeaders)
	return standard.Then(router)
}
