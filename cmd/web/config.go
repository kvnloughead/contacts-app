package main

import (
	"flag"
	"fmt"
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
	Debug BoolFlag

	// Provides verbose logging and responses in some situations. Currently only
	// middleware.logRequest makes use of this.
	Verbose BoolFlag
	DB      DatabaseConfig
}

type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

type BoolFlag struct {
	isSet bool
	value bool
}

func (b *BoolFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	b.isSet = true
	b.value = v
	return nil
}

func (b *BoolFlag) String() string {
	return fmt.Sprintf("%v", b.value)
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file:", err)
	}

	var cfg Config
	// var debug BoolFlag

	flag.IntVar(&cfg.Port, "port", 4000, "The port to run the app on.")
	flag.StringVar(&cfg.Env,
		"env",
		"development",
		"Environment (development|staging|production)")
	flag.Var(&cfg.Debug, "debug", "Run in debug mode")
	flag.Var(&cfg.Verbose, "verbose", "Provide verbose logging")

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

	if !cfg.Verbose.isSet {
		cfg.Verbose.value = os.Getenv("VERBOSE") == "true"
	}
	if !cfg.Debug.isSet {
		cfg.Debug.value = os.Getenv("DEBUG") == "true"
	}

	return cfg
}
