package reports

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
)

// Init initializes the endpoints regarding reports.
// nolint: wrapcheck,nolintlint
func Init(svc *smis.Service, db *sqlx.DB) error {
	endpoint := &reports{db: db, svc: svc}

	if _, err := svc.RegisterEndpoint("/reports/options", http.MethodGet, endpoint.options); err != nil {
		return err
	}

	_, err := svc.RegisterEndpoint("/reports/{year}", http.MethodGet, endpoint.reports)

	return err
}
