package workstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/go-utils/uuidutils"
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

// Work represents the work in the database.
type Work struct {
	ID         uuid.UUID `db:"id"`
	Start      time.Time `db:"start"`
	Stop       time.Time `db:"stop"`
	CreatedAt  time.Time `db:"created_at"`
	ModifiedAt time.Time `db:"modified_at"`
}

// Create creates current object in the database.
func (w *Work) Create(ctx context.Context, db *sqlx.DB) error {
	if !w.IsValid() {
		return ErrDataMissing
	}

	if !uuidutils.IsEmpty(w.ID) {
		return ErrIDIsSet
	}

	var err error

	w.ID, err = uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCreatingID, err)
	}

	q := db.Rebind(`
		INSERT INTO works (id, start, stop) 
		VALUES (?, ?, ?);
	`)

	_, err = db.ExecContext(ctx, q, w.ID, w.Start, w.Stop)
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}

	return w.Read(ctx, db)
}

// Read sets the work from database by given ID.
func (w *Work) Read(ctx context.Context, db *sqlx.DB) error {
	if w == nil || uuidutils.IsEmpty(w.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(`
        SELECT id, start, stop, created_at, modified_at
        FROM works
        WHERE id = ?;
    `)
	if err := db.GetContext(ctx, w, q, w.ID); err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	return nil
}

// Update changes the current object on the database by ID.
func (w *Work) Update(ctx context.Context, db *sqlx.DB) error {
	if !w.IsValid() {
		return ErrDataMissing
	}

	if uuidutils.IsEmpty(w.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(`
		UPDATE works 
		SET start = ?, stop = ? 
		WHERE id = ?;
	`)

	_, err := db.ExecContext(ctx, q, w.Start, w.Stop, w.ID)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	return w.Read(ctx, db)
}

// Delete removes the current object from database by its ID.
func (w *Work) Delete(ctx context.Context, db *sqlx.DB) error {
	if w == nil || uuidutils.IsEmpty(w.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(`
        DELETE FROM works
        WHERE id = ?
    `)

	if _, err := db.ExecContext(ctx, q, w.ID); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}

// IsValid returns true if all mandatory fields are set.
func (w *Work) IsValid() bool {
	if w == nil || w.Start.IsZero() {
		return false
	}

	return true
}
