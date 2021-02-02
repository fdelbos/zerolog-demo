package main

import (
	"database/sql"
	"os"

	"github.com/pkg/errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zerologadapter"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {

	// current logging instance
	logger := log.Logger

	// display file and line number
	logger = logger.With().Caller().Logger()

	// by default logs are in json format but we can use text format
	if os.Getenv("FORMAT") == "text" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	switch os.Getenv("LEVEL") {
	case "info":
		logger = logger.Level(zerolog.InfoLevel)
	case "warn":
		logger = logger.Level(zerolog.WarnLevel)
	case "error":
		logger = logger.Level(zerolog.ErrorLevel)
	default:
		logger = logger.Level(zerolog.DebugLevel)
	}

	// zerolog can be globally initialized
	log.Logger = logger
}

func main() {
	// required for stack trace
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// we can use it the same way we use the standard log package
	log.Print("something happend")
	log.Printf("test %d", 42)

	demoWithParams("jean jean", 38)
	sampling()
	dbLogging()

	err := outer()
	log.Error().Stack().Err(err).Msg("this is a nice stack!")

	// sub-logger
	sub := log.With().
		Str("cat name", "Spartacus").
		Str("teeth", "sharps").
		Str("heart", "big").
		Logger()
	useSubLogger(sub)
}

func demoWithParams(name string, age int) {
	err := errors.New("something is wrong!")

	// we can add more context
	log.Error().
		Err(err).
		Int("age", age).
		Str("name", name).
		Msg("an error occurred!")
}

func sampling() {
	// we can also do some sampling
	sampled := log.Sample(&zerolog.BasicSampler{N: 10})
	for i := 0; i < 30; i++ {
		sampled.Info().Msg("will be logged every 10 messages")
	}
}

func dbLogging() {
	// a standard db conn, compatible with what we are using in the api (i think...)
	var db *sql.DB

	config, err := pgx.ParseConfig("postgres://localhost/toggl_api_test")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("cant parse database connection url")
	}

	// we add a logger to the database. Note that its inherits the properties of our default logger,
	// but we could override them here (ex: level...)
	config.Logger = zerologadapter.NewLogger(log.Logger)

	db = stdlib.OpenDB(*config)
	if err := db.Ping(); err != nil {
		log.Fatal().
			Err(err).
			Msg("cant ping the database")
	}

	// a successfull request
	db.Exec("select count(*) from workspace")

	// a failed request
	db.Exec("select 'toto' = $1", 42)
}

func inner() error {
	return errors.New("seems we have an error here")
}

func middle() error {
	err := inner()
	if err != nil {
		return err
	}
	return nil
}

func outer() error {
	err := middle()
	if err != nil {
		return err
	}
	return nil
}

func useSubLogger(log zerolog.Logger) {
	err := errors.New("😻 my cat is too cute i cant look away!")
	log.Error().
		Err(err).
		Msg("this is too much for me 😱")
}
