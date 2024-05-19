package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/justinas/nosurf"
)

// Sets secure headers, per OWASP guidelines.
// https://owasp.org/www-project-secure-headers/
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

// Middleware to logs each HTTP request, including the requests IP, protocol,
// method, and URI.
//
// By default, logging of static files is suppressed for the following
// extensions: css, js, png, jpg, jpeg, ico. This behavior can be changed by
// supplying the CLI flag -verbose when running the application.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip       = r.RemoteAddr
			protocol = r.Proto
			method   = r.Method
			uri      = r.URL.RequestURI()
		)

		// Skip logging for static files, unless -verbose is true.
		if !app.config.Verbose.value {
			if strings.HasSuffix(r.URL.Path, ".css") ||
				strings.HasSuffix(r.URL.Path, ".js") ||
				strings.HasSuffix(r.URL.Path, ".png") ||
				strings.HasSuffix(r.URL.Path, ".jpg") ||
				strings.HasSuffix(r.URL.Path, ".jpeg") ||
				strings.HasSuffix(r.URL.Path, ".ico") {
				next.ServeHTTP(w, r)
				return
			}
		}

		app.logger.Info("received request", "ip", ip, "protocol", protocol, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

/*
Middleware to recover from panics and return a 500 server error. This should probably be the first middleware in the chain.

Note that this middleware will only have effect within a given go routine. So if a separate goroutine is initiated, you should include code to recover from panics inside that goroutine.

# Example

	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprint(err))
			}
		}()
		doSomeBackgroundProcessing()
	}()
*/
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/* Deferred functions are always run after a panic, as Go "unwinds" the handler stack. */
		defer func() {
			// If panicing, recover() restores normal execution and returns the error.
			if err := recover(); err != nil {
				// Set a header that will automatically close the HTTP connection after
				// response is set.
				w.Header().Set("Connection", "close")

				// Return type of recover() is any, so we need to format it as an error.
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If user isn't logged in, redirect to login page.
		if !app.isAuthenticated(r) {
			app.sessionManager.Put(r.Context(), string(redirectAfterLogin), r.URL.Path)
			app.sessionManager.Put(r.Context(), string(flash), "You must log in token access this resource.")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Prevent pages that require authorization from being cached.
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// Middleware function that uses the nosurf package to prevent CSRF attacks.
// This middleware should be used on all pages that contain a potentially
// vulnerable route (non-GET/HEAD/OPTIONS/TRACE).
//
// For this application, a logout form is included in the nav partial template
// which is included on all pages, so all non-static GET routes should be
// protected with the middleware.
//
// Additionally, CSRF token has been added to our templateData struct and
// embedded in all of our applications forms.
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}
