package reports

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
)

// Init initializes the endpoints regarding reports.
func Init(svc *smis.Service, db *sqlx.DB) error {
	endpoint := &reports{db: db, svc: svc}

	_, err := svc.RegisterEndpoint("/reports/options", http.MethodGet, endpoint.options)

	return err //nolint: wrapcheck
}
