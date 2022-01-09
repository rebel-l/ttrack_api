package timelogmapper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/go-utils/uuidutils"
	"github.com/rebel-l/ttrack_api/timelog/timelogmodel"
	"github.com/rebel-l/ttrack_api/timelog/timelogstore"
)

var (
	// ErrLoadFromDB occurs if something went wrong on loading.
	ErrLoadFromDB = errors.New("failed to load timelog from database")

	// ErrNoData occurs if given model is nil.
	ErrNoData = errors.New("timelog is nil")

	// ErrSaveToDB occurs if something went wrong on saving.
	ErrSaveToDB = errors.New("failed to save timelog to database")

	// ErrDeleteFromDB occurs if something went wrong on deleting.
	ErrDeleteFromDB = errors.New("failed to delete timelog from database")

	// ErrNotFound occurs if record doesn't exist in database.
	ErrNotFound = errors.New("timelog was not found")
)

// Mapper provides methods to load and persist timelog models.
type Mapper struct {
	db *sqlx.DB
}

// New returns a new mapper.
func New(db *sqlx.DB) *Mapper {
	return &Mapper{db: db}
}

// Load returns a timelog model loaded from database by ID.
func (m *Mapper) Load(ctx context.Context, id uuid.UUID) (*timelogmodel.Timelog, error) {
	s := &timelogstore.Timelog{ID: id} // nolint: exhaustivestruct

	if err := s.Read(ctx, m.db); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFromDB, err)
	}

	return StoreToModel(s), nil
}

// Save persists (create or update) the model and returns the changed data (id, createdAt or modifiedAt).
func (m *Mapper) Save(ctx context.Context, model *timelogmodel.Timelog) (*timelogmodel.Timelog, error) {
	if model == nil {
		return nil, ErrNoData
	}

	s := modelToStore(model)

	if uuidutils.IsEmpty(model.ID) {
		if err := s.Create(ctx, m.db); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSaveToDB, err)
		}
	} else {
		if err := s.Update(ctx, m.db); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSaveToDB, err)
		}
	}

	model = StoreToModel(s)

	return model, nil
}

// Delete removes a model from database by ID.
func (m *Mapper) Delete(ctx context.Context, id uuid.UUID) error {
	s := &timelogstore.Timelog{ID: id} // nolint: exhaustivestruct
	if err := s.Delete(ctx, m.db); err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFromDB, err)
	}

	return nil
}

// StoreToModel returns a model based on the given store object. It maps all properties from store to model.
func StoreToModel(s *timelogstore.Timelog) *timelogmodel.Timelog {
	if s == nil {
		return &timelogmodel.Timelog{} // nolint: exhaustivestruct
	}

	return &timelogmodel.Timelog{
		ID:         s.ID,
		Start:      s.Start,
		Stop:       s.Stop,
		Reason:     s.Reason,
		Location:   s.Location,
		CreatedAt:  s.CreatedAt,
		ModifiedAt: s.ModifiedAt,
	}
}

// modelToStore returns a store based on the given model object. It maps all properties from model to store.
func modelToStore(m *timelogmodel.Timelog) *timelogstore.Timelog {
	return &timelogstore.Timelog{
		ID:         m.ID,
		Start:      m.Start,
		Stop:       m.Stop,
		Reason:     m.Reason,
		Location:   m.Location,
		CreatedAt:  m.CreatedAt,
		ModifiedAt: m.ModifiedAt,
	}
}
