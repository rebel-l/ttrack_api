package timelogstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/go-utils/uuidutils"
)

const (
	qSelect = `
		SELECT id, start, stop, reason, location, created_at, modified_at
        FROM timelogs
	`
)

var (
	// ErrIDMissing will be thrown if an ID is expected but not set.
	ErrIDMissing = errors.New("id is mandatory for this operation")

	// ErrCreatingID will be thrown if creating an ID failed.
	ErrCreatingID = errors.New("id creation failed")

	// ErrIDIsSet will be thrown if no ID is expected but already set.
	ErrIDIsSet = errors.New("id should be not set for this operation, use update instead")

	// ErrDataMissing will be thrown if mandatory data is not set.
	ErrDataMissing = errors.New("no data or mandatory data missing")
)

// Timelog represents the timelog in the database.
type Timelog struct {
	ID         uuid.UUID  `db:"id"`
	Start      time.Time  `db:"start"`
	Stop       *time.Time `db:"stop"`
	Reason     string     `db:"reason"`
	Location   string     `db:"location"`
	CreatedAt  time.Time  `db:"created_at"`
	ModifiedAt time.Time  `db:"modified_at"`
}

// Create creates current object in the database.
func (t *Timelog) Create(ctx context.Context, db *sqlx.DB) error {
	if !t.IsValid() {
		return ErrDataMissing
	}

	if !uuidutils.IsEmpty(t.ID) {
		return ErrIDIsSet
	}

	var err error

	t.ID, err = uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCreatingID, err)
	}

	q := db.Rebind(`
		INSERT INTO timelogs (id, start, stop, reason, location) 
		VALUES (?, ?, ?, ?, ?);
	`)

	_, err = db.ExecContext(ctx, q, t.ID, t.Start, t.Stop, t.Reason, t.Location)
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}

	return t.Read(ctx, db)
}

// Read sets the timelog from database by given ID.
func (t *Timelog) Read(ctx context.Context, db *sqlx.DB) error {
	if t == nil || uuidutils.IsEmpty(t.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(
		qSelect + `
        WHERE id = ?;
    `)
	if err := db.GetContext(ctx, t, q, t.ID); err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	return nil
}

// Update changes the current object on the database by ID.
func (t *Timelog) Update(ctx context.Context, db *sqlx.DB) error {
	if !t.IsValid() {
		return ErrDataMissing
	}

	if uuidutils.IsEmpty(t.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(`
		UPDATE timelogs 
		SET start = ?, stop = ?, reason = ?, location = ? 
		WHERE id = ?;
	`)

	_, err := db.ExecContext(ctx, q, t.Start, t.Stop, t.Reason, t.Location, t.ID)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	return t.Read(ctx, db)
}

// Delete removes the current object from database by its ID.
func (t *Timelog) Delete(ctx context.Context, db *sqlx.DB) error {
	if t == nil || uuidutils.IsEmpty(t.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(`
        DELETE FROM timelogs
        WHERE id = ?
    `)

	if _, err := db.ExecContext(ctx, q, t.ID); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}

// IsValid returns true if all mandatory fields are set.
func (t *Timelog) IsValid() bool {
	if t == nil || t.Start.IsZero() || t.Reason == "" || t.Location == "" {
		return false
	}

	return true
}
