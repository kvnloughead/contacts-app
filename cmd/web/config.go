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

// DatabaseConfig is a struct that stores database configuration. The DSN field
// will be necessary to connect to the database, and will be pulled from a .env
// file if there is one.
type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

// BoolFlag is a struct to store boolean flags. It implements the Set method
// which is called when the flags are parsed. If a flag has been passed at the
// command line the isSet field will be set to true. This can be used to
// distinguish between a default 'false' value and an unset flag.
type BoolFlag struct {
	// If isSet is false, the flag has not been set.
	isSet bool

	// The value of the flag. If isSet is false, then this will be the default.
	value bool
}

// The Set method is called whenever flag.Parse is called. If the string
// argument can be converted into a bool, then this bool is set as the
// BoolFlag's value and isSet is set to true.
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

// loadIntFromEnvOrFlag loads an integer valued config option and assigns it to
// the target int. This function should be called after flags are parsed with
// flag.Parse.
//
// It then checks if the target has the default value. If not, no action is
// taken, because the flags should override environmental variables. If it still
// has the default value, the function checks for a matching environmental
// variable. If it exists and can be converted into an integer, it is assigned
// to the target.
func loadIntFromEnvOrFlag(target *int, defaultVal int, envKey string) {
	if *target == defaultVal {
		if envVar, ok := os.LookupEnv(envKey); ok {
			val, err := strconv.Atoi(envVar)
			if err == nil {
				*target = val
			}
		}
	}
}

// loadDurationFromEnvOrFlag loads a time.Duration valued config option and
// assigns it to the target. This function should be called after flags are
// parsed with flag.Parse.
//
// It then checks if the target has the default value. If not, no action is
// taken, because the flags should override environmental variables. If it still
// has the default value, the function checks for a matching environmental
// variable. If it exists and can be converted into a time.Duration, it is
// assigned to the target.
func loadDurationFromEnvOrFlag(target *time.Duration, defaultVal time.Duration, envKey string) {
	if *target == defaultVal {
		if envVar, ok := os.LookupEnv(envKey); ok {
			val, err := time.ParseDuration(envVar)
			if err == nil {
				*target = val
			}
		}
	}
}

// LoadConfig loads the configuration, returning the resulting Config struct.
// It first loads environmental variables from the environment, including from
// a .env file. Then, if any command line flags have been set, these will
// override the evironmental variables.
//
// Reasonable defaults have been provided in most cases. The exception is the
// -db-dsn flag, which defaults to an empty string. This must be provided as
// an environmental variable or flag.
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
	flag.Var(&cfg.Debug, "debug", "Run in debug mode")
	flag.Var(&cfg.Verbose, "verbose", "Provide verbose logging")

	// Read DB-related settings from CLI flags.
	flag.StringVar(&cfg.DB.DSN, "db-dsn", "", "Postgresql DSN")
	flag.IntVar(&cfg.DB.MaxOpenConns, "db-max-open-conns", 25, "Postgresql max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConns, "db-max-idle-conns", 25, "Postgresql max idle connections")
	flag.DurationVar(&cfg.DB.MaxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgresql max connection idle time")

	flag.Parse()

	// If DSN is not supplied by flag, load it from the environment. The DSN is
	// required.
	if cfg.DB.DSN == "" {
		cfg.DB.DSN = os.Getenv("DB_DSN")
	}

	// Load integer and duration valued configuration options.
	loadIntFromEnvOrFlag(&cfg.Port, 4000, "PORT")
	loadIntFromEnvOrFlag(&cfg.DB.MaxOpenConns, 25, "DB_MAX_OPEN_CONNS")
	loadIntFromEnvOrFlag(&cfg.DB.MaxIdleConns, 25, "DB_MAX_IDLE_CONNS")
	loadDurationFromEnvOrFlag(&cfg.DB.MaxIdleTime, 15*time.Minute, "DB_MAX_IDLE_TIME")

	// Load Boolean valued configuration options.
	if !cfg.Verbose.isSet {
		cfg.Verbose.value = os.Getenv("VERBOSE") == "true"
	}
	if !cfg.Debug.isSet {
		cfg.Debug.value = os.Getenv("DEBUG") == "true"
	}

	return cfg
}
