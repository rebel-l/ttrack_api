package publicholidaystore

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
        SELECT id, day, name, halfday, created_at, modified_at
        FROM publicholidays
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

// PublicHoliday represents the publicholiday in the database.
type PublicHoliday struct {
	ID         uuid.UUID `db:"id"`
	Day        time.Time `db:"day"`
	Name       string    `db:"name"`
	HalfDay    bool      `db:"halfday"`
	CreatedAt  time.Time `db:"created_at"`
	ModifiedAt time.Time `db:"modified_at"`
}

// Create creates current object in the database.
func (p *PublicHoliday) Create(ctx context.Context, db *sqlx.DB) error {
	if !p.IsValid() {
		return ErrDataMissing
	}

	if !uuidutils.IsEmpty(p.ID) {
		return ErrIDIsSet
	}

	var err error

	p.ID, err = uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCreatingID, err)
	}

	q := db.Rebind(`
		INSERT INTO publicholidays (id, day, name, halfday) 
		VALUES (?, ?, ?, ?);
	`)

	_, err = db.ExecContext(ctx, q, p.ID, p.Day, p.Name, p.HalfDay)
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}

	return p.Read(ctx, db)
}

// Read sets the publicholiday from database by given ID.
func (p *PublicHoliday) Read(ctx context.Context, db *sqlx.DB) error {
	if p == nil || uuidutils.IsEmpty(p.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(
		qSelect + `
        WHERE id = ?;
    `)

	if err := db.GetContext(ctx, p, q, p.ID); err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	return nil
}

// Update changes the current object on the database by ID.
func (p *PublicHoliday) Update(ctx context.Context, db *sqlx.DB) error {
	if !p.IsValid() {
		return ErrDataMissing
	}

	if uuidutils.IsEmpty(p.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(`
		UPDATE publicholidays 
		SET day = ?, name = ?, halfday = ? 
		WHERE id = ?;
	`)

	_, err := db.ExecContext(ctx, q, p.Day, p.Name, p.HalfDay, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	return p.Read(ctx, db)
}

// Delete removes the current object from database by its ID.
func (p *PublicHoliday) Delete(ctx context.Context, db *sqlx.DB) error {
	if p == nil || uuidutils.IsEmpty(p.ID) {
		return ErrIDMissing
	}

	q := db.Rebind(`
        DELETE FROM publicholidays
        WHERE id = ?
    `)

	if _, err := db.ExecContext(ctx, q, p.ID); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}

// IsValid returns true if all mandatory fields are set.
func (p *PublicHoliday) IsValid() bool {
	if p == nil || p.Day.IsZero() || p.Name == "" {
		return false
	}

	return true
}
