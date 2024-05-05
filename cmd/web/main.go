package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/kvnloughead/contacts-app/internal/models"

	// Aliasing with a blank identifier because the driver isn't used explicitly.
	_ "github.com/lib/pq"
)

// A struct containing application-wide dependencies.
type application struct {
	logger         *slog.Logger
	contacts       models.ContactModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	debug          bool
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP Network Address")
	dsn := flag.String(
		"dsn",
		"web:devpass@/snippetbox?parseTime=true",
		"MySQL data source name (aka 'connection string')")
	debug := flag.Bool("debug", false, "Run in debug mode")
	flag.Parse()

	// Initialize structured logger to stdout with default settings.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true, // include file and line number
	}))

	// Initialize sql.DB connection pool for the provided DSN.
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Initialize template cache.
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize session manager, using our db as its store. We then add it to
	// our dependency injector, and wrap our routes in its LoadAndSave middleware.
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true // only send cookies over HTTPS

	formDecoder := form.NewDecoder()

	app := &application{
		logger:         logger,
		contacts:       &models.ContactModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		debug:          *debug,
	}

	// Struct containing non-default TLS settings.
	tlsConfig := tls.Config{
		// For performance, only use curves with assembly implementations.
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initial http server with address route handler.
	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		TLSConfig:    &tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,

		// Instruct our http server to log error using our structured logger.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	/* Info level log statement. Arguments after the first can either be variadic, key/value pairs, or attribute pairs created by slog.String, or a similar method. */
	logger.Info("starting server", slog.String("addr", srv.Addr))

	// Run the HTTPS server, passing it the self-signed TLS certificate and key.
	// If an error occurs, log it and exit.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

// Returns an sql.DB connection pool for the supplied data source name (DSN).
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Verify that the connection is alive, reestablishing it if necessary.
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
