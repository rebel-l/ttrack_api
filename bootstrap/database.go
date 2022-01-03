package bootstrap

import (
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/go-utils/osutils"
	"github.com/rebel-l/schema"
	"github.com/rebel-l/ttrack_api/config"
)

const (
	storageFileName = "ttrack_api.db"
)

// Database initialises the database and returns the connection.
func Database(conf *config.Database, version string, withBar bool) (*sqlx.DB, error) {
	fileName, err := createStorage(conf.GetStoragePath())
	if err != nil {
		return nil, fmt.Errorf("bootstrap database, create storage failed: %w", err)
	}

	db, err := open(fileName)
	if err != nil {
		return nil, fmt.Errorf("bootstrap database, open database failed: %w", err)
	}

	err = createSchema(db, conf.GetSchemaScriptPath(), version, withBar)
	if err != nil {
		return nil, fmt.Errorf("bootstrap database, create schema failed: %w", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, fmt.Errorf("bootstrap database, activate foreign key checks failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("bootstrap database, failed to reach database: %w", err)
	}

	return db, nil
}

func createStorage(path string) (string, error) {
	if err := osutils.CreateDirectoryIfNotExists(path); err != nil {
		return "", err // nolint: wrapcheck
	}

	fileName := buildFileName(path)
	if err := osutils.CreateFileIfNotExists(fileName); err != nil {
		return "", err // nolint: wrapcheck
	}

	return fileName, nil
}

func createSchema(db *sqlx.DB, scriptPath, version string, withBar bool) error {
	s := schema.New(db)
	if withBar {
		s.WithProgressBar()
	}

	return s.Upgrade(scriptPath, version) // nolint: wrapcheck
}

func open(fileName string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", fileName)

	return db, err // nolint: wrapcheck
}

func buildFileName(path string) string {
	return filepath.Join(path, storageFileName)
}

// DatabaseReset resets the whole database. NOTE: all data will be lost, should be used only for development.
func DatabaseReset(conf *config.Database, withBar bool) error {
	fileName := buildFileName(conf.GetStoragePath())

	db, err := open(fileName)
	if err != nil {
		return fmt.Errorf("bootstrap database reset, open database failed: %w", err)
	}

	defer func() {
		_ = db.Close()
	}()

	s := schema.New(db)
	if withBar {
		s.WithProgressBar()
	}

	err = s.RevertAll(conf.GetSchemaScriptPath())
	if err != nil {
		return fmt.Errorf("bootstrap database reset, revert database failed: %w", err)
	}

	return nil
}
