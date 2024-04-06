package timelogstore

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type UniqueYears []string

func (u *UniqueYears) Get(ctx context.Context, db *sqlx.DB) error {
	q := db.Rebind(`
			SELECT DISTINCT strftime('%Y',start)
			FROM timelogs
			WHERE start IS NOT NULL
		UNION
			SELECT DISTINCT strftime('%Y',stop)
			FROM timelogs
			WHERE stop IS NOT NULL;
	`)

	if err := db.SelectContext(ctx, u, q); err != nil {
		return fmt.Errorf("failed to load unique years: %w", err)
	}

	return nil
}
