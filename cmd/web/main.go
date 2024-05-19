package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/kvnloughead/contacts-app/internal/models"

	// Aliasing with a blank identifier because the driver isn't used explicitly.
	_ "github.com/lib/pq"
)

// A struct containing application-wide dependencies.
type application struct {
	config         Config
	logger         *slog.Logger
	contacts       models.ContactModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	var cfg = LoadConfig()

	// Initialize structured logger to stdout with default settings.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true, // include file and line number
	}))

	// Initialize sql.DB connection pool for the provided DSN.
	db, err := openDB(cfg.DB)
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
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	formDecoder := form.NewDecoder()

	app := &application{
		logger:         logger,
		contacts:       &models.ContactModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		config:         cfg,
	}

	// Initial http server with address route handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,

		// Instruct our http server to log error using our structured logger.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	/* Info level log statement. Arguments after the first can either be variadic, key/value pairs, or attribute pairs created by slog.String, or a similar method. */
	logger.Info("starting server", slog.String("port", fmt.Sprint(cfg.Port)))

	// Run the server. If an error occurs, log it and exit.
	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB returns an postgres sql.DB connection pool for the supplied DSN. It
// accepts a configuration struct as an argument, using its fields to set the
// DSN and other settings.
func openDB(dbCfg DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbCfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(dbCfg.MaxOpenConns)
	db.SetMaxIdleConns(dbCfg.MaxIdleConns)
	db.SetConnMaxIdleTime(dbCfg.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify that the connection is alive, reestablishing it if necessary.
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
