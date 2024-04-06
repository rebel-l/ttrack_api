package reports

import (
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
)

type reports struct {
	db  *sqlx.DB
	svc *smis.Service
}
