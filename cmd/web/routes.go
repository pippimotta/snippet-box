package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice" //use this package to avoid writing handler chain
	"github.com/pippimotta/snippet-box/ui"
)

func (a *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	//create a new middleware chain containing the middleware to control dynamic routes
	//divide routes into two groups, one for "protected", one for "unprotected"
	dynamic := alice.New(a.sessionManager.LoadAndSave, a.noSurf, a.authenticated)
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(a.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(a.snippetView))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(a.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(a.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(a.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(a.userLoginPost))

	protected := dynamic.Append(a.requireAuthentication)
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(a.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(a.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(a.userLogoutPost))

	standard := alice.New(a.recoverPanic, a.logRequest, secureHeaders)
	return standard.Then(router)
}
