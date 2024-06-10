package publicholidaymapper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/go-utils/uuidutils"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaystore"
)

var (
	// ErrLoadFromDB occurs if something went wrong on loading.
	ErrLoadFromDB = errors.New("failed to load publicholiday from database")

	// ErrNoData occurs if given model is nil.
	ErrNoData = errors.New("publicholiday is nil")

	// ErrSaveToDB occurs if something went wrong on saving.
	ErrSaveToDB = errors.New("failed to save publicholiday to database")

	// ErrDeleteFromDB occurs if something went wrong on deleting.
	ErrDeleteFromDB = errors.New("failed to delete publicholiday from database")

	// ErrNotFound occurs if record doesn't exist in database.
	ErrNotFound = errors.New("publicholiday was not found")

	// ErrConvert occurs if data type conversion failed.
	ErrConvert = errors.New("conversion error")
)

// Mapper provides methods to load and persist publicholiday models.
type Mapper struct {
	db *sqlx.DB
}

// New returns a new mapper.
func New(db *sqlx.DB) *Mapper {
	return &Mapper{db: db}
}

// Load returns a publicholiday model loaded from database by ID.
func (m *Mapper) Load(ctx context.Context, id uuid.UUID) (*publicholidaymodel.PublicHoliday, error) {
	s := &publicholidaystore.PublicHoliday{ID: id} // nolint: exhaustivestruct

	if err := s.Read(ctx, m.db); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFromDB, err)
	}

	return StoreToModel(s), nil
}

// Save persists (create or update) the model and returns the changed data (id, createdAt or modifiedAt).
func (m *Mapper) Save(ctx context.Context, model *publicholidaymodel.PublicHoliday) (*publicholidaymodel.PublicHoliday, error) {
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
	s := &publicholidaystore.PublicHoliday{ID: id} // nolint: exhaustivestruct
	if err := s.Delete(ctx, m.db); err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFromDB, err)
	}

	return nil
}

// StoreToModel returns a model based on the given store object. It maps all properties from store to model.
func StoreToModel(s *publicholidaystore.PublicHoliday) *publicholidaymodel.PublicHoliday {
	if s == nil {
		return &publicholidaymodel.PublicHoliday{} // nolint: exhaustivestruct
	}

	return &publicholidaymodel.PublicHoliday{
		ID:         s.ID,
		Day:        s.Day,
		Name:       s.Name,
		HalfDay:    s.HalfDay,
		CreatedAt:  s.CreatedAt,
		ModifiedAt: s.ModifiedAt,
	}
}

// modelToStore returns a store based on the given model object. It maps all properties from model to store.
func modelToStore(m *publicholidaymodel.PublicHoliday) *publicholidaystore.PublicHoliday {
	return &publicholidaystore.PublicHoliday{
		ID:         m.ID,
		Day:        m.Day,
		Name:       m.Name,
		HalfDay:    m.HalfDay,
		CreatedAt:  m.CreatedAt,
		ModifiedAt: m.ModifiedAt,
	}
}
