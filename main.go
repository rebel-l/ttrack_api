/*
Backend Service of a time tracking application.
*/
package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rebel-l/go-utils/slice"
	"github.com/rebel-l/smis"
	"github.com/rebel-l/smis/middleware/cors"
	"github.com/rebel-l/ttrack_api/bootstrap"
	"github.com/rebel-l/ttrack_api/config"
	"github.com/rebel-l/ttrack_api/endpoint/doc"
	"github.com/rebel-l/ttrack_api/endpoint/ping"
	"github.com/rebel-l/ttrack_api/endpoint/timelogs"
	"github.com/sirupsen/logrus"
)

const (
	defaultPort               = 3000
	defaultTimeout            = 15 * time.Second
	defaultAccesControlMaxAge = 86400 // 24 Hours
	version                   = "0.1.0"
)

var (
	db   *sqlx.DB
	log  logrus.FieldLogger
	port *int
	svc  *smis.Service
)

func initCustomFlags() {
	/**
	  1. Add your custom service flags below, for more details see https://golang.org/pkg/flag/
	*/
}

func initCustom() error {
	/**
	  2. add your custom service initialisation below, e.g. database connection, caches etc.
	*/
	var err error

	// Database
	db, err = bootstrap.Database(&config.Database{}, version, true) // nolint: exhaustivestruct
	if err != nil {
		return fmt.Errorf("bootstrapping database failed: %w", err)
	}

	return nil
}

func initCustomRoutes() error {
	c := cors.Config{
		AccessControlAllowOrigins: slice.StringSlice{"*"},
		AccessControlAllowHeaders: slice.StringSlice{"*"},
		AccessControlMaxAge:       defaultAccesControlMaxAge,
	} // TODO: make it configurable
	svc.WithDefaultMiddleware(c) // TODO: add catch panic middleware to default

	/**
	  3. Register your custom routes below
	*/
	if err := timelogs.Init(svc, db); err != nil {
		return fmt.Errorf("failed to init the timelogs endpoints: %w", err)
	}

	return nil
}

func main() {
	log = logrus.New()
	log.Info("Starting service: ttrack_api")

	initFlags()
	initService()

	if err := initCustom(); err != nil {
		log.Fatalf("Failed to initialise custom settings: %s", err)
	}

	if err := initRoutes(); err != nil {
		log.Fatalf("Failed to initialise routes: %s", err)
	}

	log.Infof("Service listens to port %d", *port)
	if err := svc.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}

func initService() {
	router := mux.NewRouter()
	srv := &http.Server{ // nolint: exhaustivestruct
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", *port),
		WriteTimeout: defaultTimeout,
		ReadTimeout:  defaultTimeout,
	}

	var err error
	svc, err = smis.NewService(srv, router, log)
	if err != nil {
		log.Fatalf("failed to initialize service: %s", err)
	}
}

func initRoutes() error {
	if err := initDefaultRoutes(); err != nil {
		return fmt.Errorf("default routes failed: %w", err)
	}

	if err := initCustomRoutes(); err != nil {
		return fmt.Errorf("custom routes failed: %w", err)
	}

	return nil
}

func initDefaultRoutes() error {
	if err := ping.Init(svc); err != nil {
		return fmt.Errorf("failed to init endpoint /ping: %w", err)
	}

	if err := doc.Init(svc); err != nil {
		return fmt.Errorf("failed to init endpoint /doc: %w", err)
	}

	return nil
}

func initFlags() {
	initDefaultFlags()
	initCustomFlags()
	flag.Parse()
}

func initDefaultFlags() {
	port = flag.Int("p", defaultPort, "the port the service listens to")
}
