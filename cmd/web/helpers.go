// Application helper methods.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

/*
Writes an Error level log entry, including the request method and uri, and
returns a 500 Internal Server Error.

If run in debug mode, the full stack trace is sent to the client.
*/
func (app *application) serverError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	// Log error with stack trace.
	app.logger.Error(err.Error(), "method", method, "uri", uri)

	if app.debug {
		body := fmt.Sprintf("%s\n%s", err, trace)
		http.Error(w, body, http.StatusInternalServerError)
		return
	}

	// Send http error response to client.
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

}

/*
Sends a status code and corresponding description to the user. Uses
http.StatusText to generate the standard description.
*/
func (app *application) clientError(
	w http.ResponseWriter,
	status int,
) {
	http.Error(w, http.StatusText(status), status)
}

// A wrapper around clientError to easily send 404 Not Found responses.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// Retrieves appropriate template from the cache, based on the page string.
// If no entry exists in the cache, a 500 server error is returned.
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// Write template to a buffer, instead of immediately to the ResponseWriter.
	// If there's an error, return a server error instead of a 200 response.
	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// If template is written to buffer with no errors, we are safe to send a 200
	// response to the client.
	w.WriteHeader(status)
	buf.WriteTo(w) // write contents of buffer to http.ResponseWriter
}

// Initialize templateData struct with the CurrentYear.
func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), string(flash)),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
	}
}

/* Parses form with r.ParseForm() and then attempts to decode r.PostForm into the target destination dst. If dst is invalid, an InvalidDecodeError occurs, a panic ensues. Otherwise, the error is returned to the caller. */
func (app *application) decodePostForm(r *http.Request, dst any) error {
	// r.ParseForm() populates r.Form and r.PostForm and validates response body.
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecodeError *form.InvalidDecoderError
		if errors.As(err, &invalidDecodeError) {
			panic(err)
		}
		return err
	}

	return nil
}

// Returns true if the request is coming from an authenticated user. Authentication is determined by the presence and value of an isAuthenticatedContextKey in the request context.
//
// False will be returned if the key doesn't exist, if it's value isn't boolean, or if its value is false.
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}
