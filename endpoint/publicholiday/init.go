package publicholiday

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
)

// Init initializes the endpoints regarding publicholiday.
// nolint: wrapcheck,nolintlint
func Init(svc *smis.Service, db *sqlx.DB) error {
	endpoint := &publicHoliday{db: db, svc: svc}

	if _, err := svc.RegisterEndpoint("/publicholidays", http.MethodGet, endpoint.loadAll); err != nil {
		return err
	}

	_, err := svc.RegisterEndpoint("/publicholidays", http.MethodPut, endpoint.save)

	return err
}
