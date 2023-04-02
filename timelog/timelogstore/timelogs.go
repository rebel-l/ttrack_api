package timelogstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Timelogs []*Timelog

func (t *Timelogs) Load(ctx context.Context, db *sqlx.DB, where string, args ...any) error {
	q := qSelect
	if where != "" {
		q += " WHERE " + where
	}

	if err := db.SelectContext(ctx, t, db.Rebind(q), args...); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	return nil
}
