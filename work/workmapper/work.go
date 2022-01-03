package workmapper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/go-utils/uuidutils"
	"github.com/rebel-l/ttrack_api/work/workmodel"
	"github.com/rebel-l/ttrack_api/work/workstore"
)

var (
	// ErrLoadFromDB occurs if something went wrong on loading.
	ErrLoadFromDB = errors.New("failed to load work from database")

	// ErrNoData occurs if given model is nil.
	ErrNoData = errors.New("work is nil")

	// ErrSaveToDB occurs if something went wrong on saving.
	ErrSaveToDB = errors.New("failed to save work to database")

	// ErrDeleteFromDB occurs if something went wrong on deleting.
	ErrDeleteFromDB = errors.New("failed to delete work from database")

	// ErrNotFound occurs if record doesn't exist in database.
	ErrNotFound = errors.New("work was not found")
)

// Mapper provides methods to load and persist work models.
type Mapper struct {
	db *sqlx.DB
}

// New returns a new mapper.
func New(db *sqlx.DB) *Mapper {
	return &Mapper{db: db}
}

// Load returns a work model loaded from database by ID.
func (m *Mapper) Load(ctx context.Context, id uuid.UUID) (*workmodel.Work, error) {
	s := &workstore.Work{ID: id} // nolint: exhaustivestruct

	if err := s.Read(ctx, m.db); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFromDB, err)
	}

	return StoreToModel(s), nil
}

// Save persists (create or update) the model and returns the changed data (id, createdAt or modifiedAt).
func (m *Mapper) Save(ctx context.Context, model *workmodel.Work) (*workmodel.Work, error) {
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
	s := &workstore.Work{ID: id} //nolint: exhaustivestruct
	if err := s.Delete(ctx, m.db); err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFromDB, err)
	}

	return nil
}

// StoreToModel returns a model based on the given store object. It maps all properties from store to model.
func StoreToModel(s *workstore.Work) *workmodel.Work {
	if s == nil {
		return &workmodel.Work{} // nolint: exhaustivestruct
	}

	return &workmodel.Work{
		ID:         s.ID,
		Start:      s.Start,
		Stop:       s.Stop,
		CreatedAt:  s.CreatedAt,
		ModifiedAt: s.ModifiedAt,
	}
}

// modelToStore returns a store based on the given model object. It maps all properties from model to store.
func modelToStore(m *workmodel.Work) *workstore.Work {
	return &workstore.Work{
		ID:         m.ID,
		Start:      m.Start,
		Stop:       m.Stop,
		CreatedAt:  m.CreatedAt,
		ModifiedAt: m.ModifiedAt,
	}
}
