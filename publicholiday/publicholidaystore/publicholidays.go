package publicholidaystore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type PublicHolidays []*PublicHoliday

func (p *PublicHolidays) Load(ctx context.Context, db *sqlx.DB, where string, args ...any) error {
	q := qSelect
	if where != "" {
		q += " WHERE " + where
	}
	q += " ORDER BY day "

	if err := db.SelectContext(ctx, p, db.Rebind(q), args...); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	return nil
}
