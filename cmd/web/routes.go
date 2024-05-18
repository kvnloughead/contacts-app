package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"

	"github.com/kvnloughead/contacts-app/ui"
)

/*
Returns a servemux that serves files from ./ui/static and contains the following routes:

Static unprotected routes
  - GET  /static/*filepath    serve a static file

Dynamic unprotected routes:
  - GET  		/											   			display the home page
  - GET  		/about												display the about page
  - GET  		/ping 							  				responses with 200 OK
  - GET  		/contacts/create   	   		    display form to create contacts
  - POST 		/contacts/create      				create a new contact
  - GET  		/contacts/view/:id        		display a specific contact
  - GET  		/contacts/edit/:id        		display edit form a contact
  - POST 		/contacts/edit/:id        		edit a contact
  - GET     /contacts/delete/:id          display contact and prompts to delete
  - POST    /contacts/delete/:id          delete a contact

Currently all HTTP requests are GET or POST. I intend to change this with
HTMX at a later time.
*/
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Use our app.notFound method instead of httprouter's built-in 404 handler.
	router.NotFound = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			app.notFound(w)
		})

	// Serve static files out of embedded filesystem ui.Files.
	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(
		http.MethodGet,
		"/static/*filepath",
		fileServer,
	)

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// Middleware chain for dynamic routes only (not static files).
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf)

	// Dynamic routes are wrapped in our dynamic middleware. Note that since
	// ThenFunc returns an http.Handler, we need to use router.Handler instead of
	// router.HandlerFunc.
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/about", dynamic.ThenFunc(app.about))
	router.Handler(http.MethodGet, "/contacts/view/:id", dynamic.ThenFunc(app.contactView))

	router.Handler(http.MethodGet, "/contacts/edit/:id", dynamic.ThenFunc(app.contactEdit))
	router.Handler(http.MethodPost, "/contacts/edit/:id", dynamic.ThenFunc(app.contactEditPost))

	router.Handler(http.MethodGet, "/contacts/delete/:id", dynamic.ThenFunc(app.contactDelete))
	router.Handler(http.MethodPost, "/contacts/delete/:id", dynamic.ThenFunc(app.contactDeletePost))

	router.Handler(http.MethodGet, "/contacts/create", dynamic.ThenFunc(app.contactCreate))
	router.Handler(http.MethodPost, "/contacts/create", dynamic.ThenFunc(app.contactCreatePost))

	// Initialize chain of standard pre-request middlewares.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
