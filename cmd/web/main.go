package main

import (
	"database/sql"
	"flag"
	"fmt"
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

// config is a struct containing configuration settings. These settings are
// specified as CLI flags when application starts, and have defaults provided
// in case they are omitted.
type config struct {
	port  int
	env   string
	debug bool
	db    struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

// A struct containing application-wide dependencies.
type application struct {
	config         config
	logger         *slog.Logger
	contacts       models.ContactModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "The port to run the app on.")
	flag.StringVar(&cfg.env,
		"env",
		"development",
		"Environment (development|staging|production)")
	flag.BoolVar(&cfg.debug, "debug", false, "Run in debug mode")

	// Read DB-related settings from CLI flags.
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "Postgresql DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Postgresql max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Postgresql max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgresql max connection idle time")

	flag.Parse()

	// Initialize structured logger to stdout with default settings.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true, // include file and line number
	}))

	// Initialize sql.DB connection pool for the provided DSN.
	db, err := openDB(cfg.db.dsn)
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

	formDecoder := form.NewDecoder()

	app := &application{
		logger:         logger,
		contacts:       &models.ContactModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Initial http server with address route handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,

		// Instruct our http server to log error using our structured logger.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	/* Info level log statement. Arguments after the first can either be variadic, key/value pairs, or attribute pairs created by slog.String, or a similar method. */
	logger.Info("starting server", slog.String("addr", srv.Addr))

	// Run the server. If an error occurs, log it and exit.
	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// Returns an sql.DB connection pool for the supplied data source name (DSN).
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
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
