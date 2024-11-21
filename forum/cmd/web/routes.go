package main

import (
	"net/http"

	"github.com/justinas/alice"
)

// routes returns a mux with all the registered routes.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	mux.Handle("GET /{$}", dynamic.ThenFunc(http.HandlerFunc(app.home)))
	mux.Handle("GET /account/create", dynamic.ThenFunc(app.accountCreate))
	mux.Handle("POST /account/create", dynamic.ThenFunc(app.accountCreatePOST))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))
	mux.Handle("GET /account/view/{id}", protected.ThenFunc(app.accountView))
	mux.Handle("GET /thread/create", protected.ThenFunc(app.threadCreate))
	mux.Handle("POST /thread/create", protected.ThenFunc(app.threadCreatePOST))
	mux.Handle("GET /thread/view/{id}", protected.ThenFunc(app.threadView))
	mux.Handle("GET /thread/view/{id}/post/create", protected.ThenFunc(app.postCreate))
	mux.Handle("POST /thread/view/{id}/post/create", protected.ThenFunc(app.postCreatePOST))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	standard := alice.New(app.logRequest, commonHeaders)
	return standard.Then(mux)
}
