package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// config is a struct containing configuration settings. These settings are
// specified as CLI flags when application starts, and have defaults provided
// in case they are omitted.
type Config struct {
	Port int
	Env  string

	// Sends full stack trace of server errors in response.
	Debug bool

	// Provides verbose logging and responses in some situations. Currently only
	// middleware.logRequest makes use of this.
	Verbose bool
	DB      DatabaseConfig
}

type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file:", err)
	}

	var cfg Config

	flag.IntVar(&cfg.Port, "port", 4000, "The port to run the app on.")
	flag.StringVar(&cfg.Env,
		"env",
		"development",
		"Environment (development|staging|production)")
	flag.BoolVar(&cfg.Debug, "debug", false, "Run in debug mode")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Provide verbose logging")

	// Read DB-related settings from CLI flags.
	flag.StringVar(&cfg.DB.DSN, "db-dsn", "", "Postgresql DSN")
	flag.IntVar(&cfg.DB.MaxOpenConns, "db-max-open-conns", 25, "Postgresql max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConns, "db-max-idle-conns", 25, "Postgresql max idle connections")
	flag.DurationVar(&cfg.DB.MaxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgresql max connection idle time")

	flag.Parse()

	// Check for environmental variables
	if cfg.DB.DSN == "" {
		cfg.DB.DSN = os.Getenv("CONTACTS_DB_DSN")
	}
	if cfg.Port == 4000 {
		if portEnv, ok := os.LookupEnv("PORT"); ok {
			port, err := strconv.Atoi(portEnv)
			if err == nil {
				cfg.Port = port
			}
		}
	}
	if !cfg.Verbose {
		cfg.Verbose = os.Getenv("VERBOSE") == "true"
	}
	if !cfg.Debug {
		cfg.Debug = os.Getenv("DEBUG") == "true"
	}

	return cfg
}
