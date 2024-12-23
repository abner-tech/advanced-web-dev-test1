package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/abner-tech/Test1/internal/data"
	_ "github.com/lib/pq"
)

const appVersion = "3.0.0"

type serverConfig struct {
	port        int
	environment string
	db          struct {
		dsn string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type applicationDependences struct {
	config       serverConfig
	logger       *slog.Logger
	productModel data.ProductModel
	reviewModel  data.ReviewModel
}

func main() {
	var settings serverConfig
	flag.IntVar(&settings.port, "port", 4000, "Server Port")
	flag.StringVar(&settings.environment, "env", "development", "Environment(development|staging|production)")
	//read the dsn
	flag.StringVar(&settings.db.dsn, "db-dsn", "postgres://test1:test1@localhost/test1?sslmode=disable", "PostgreSQL DSN")
	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")
	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")
	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	//the call to openDB() sets up our connection pool
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	//release the database connection before exiting
	defer db.Close()

	logger.Info("Database Connection Pool Established")

	appInstance := &applicationDependences{
		config:       settings,
		logger:       logger,
		productModel: data.ProductModel{DB: db},
		reviewModel:  data.ReviewModel{DB: db},
	}

	// apiServer := &http.Server{
	// 	Addr:         fmt.Sprintf(":%d", settings.port),
	// 	Handler:      appInstance.routes(),
	// 	IdleTimeout:  time.Minute,
	// 	ReadTimeout:  5 * time.Second,
	// 	WriteTimeout: 10 * time.Second,
	// 	ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	// }

	// logger.Info("Starting Server", "address", apiServer.Addr, "environment", settings.environment, "rateLimiter", settings.limiter)
	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(settings serverConfig) (*sql.DB, error) {
	//open a connection pool
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}

	//set context to ensure DB operations dont take too long
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	//pinging connection pool to verify it was created, with a 5-second timeout
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	//return the connection pool (sql.DB)
	return db, nil
}
