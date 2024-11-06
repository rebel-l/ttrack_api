package timelogs

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
)

// Init initializes the endpoints to log times.
func Init(svc *smis.Service, db *sqlx.DB) error {
	endpoint := &timelog{db: db, svc: svc}

	if _, err := svc.RegisterEndpoint("/timgelogs", http.MethodPut, endpoint.upsert); err != nil {
		return err
	}

	if _, err := svc.RegisterEndpoint("/timgelogs/{id}", http.MethodDelete, endpoint.delete); err != nil {
		return err
	}

	_, err := svc.RegisterEndpoint("/timelogs/{start}/{stop}", http.MethodGet, endpoint.loadByRange)

	return err // nolint: wrapcheck
}
