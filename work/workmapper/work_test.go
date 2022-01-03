package workmapper_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rebel-l/go-utils/osutils"
	"github.com/rebel-l/go-utils/testingutils"
	"github.com/rebel-l/ttrack_api/bootstrap"
	"github.com/rebel-l/ttrack_api/config"
	"github.com/rebel-l/ttrack_api/work/workmapper"
	"github.com/rebel-l/ttrack_api/work/workmodel"
	"github.com/rebel-l/ttrack_api/work/workstore"
)

func setup(t *testing.T, name string) *sqlx.DB {
	t.Helper()

	// 0. init path
	storagePath := filepath.Join(".", "..", "..", "storage", "test_work", name)
	scriptPath := filepath.Join(".", "..", "..", "scripts", "sql", "sqlite")
	conf := &config.Database{
		StoragePath:       &storagePath,
		SchemaScriptsPath: &scriptPath,
	}

	// 1. clean up
	if osutils.FileOrPathExists(conf.GetStoragePath()) {
		if err := os.RemoveAll(conf.GetStoragePath()); err != nil {
			t.Fatalf("failed to cleanup test files: %v", err)
		}
	}

	// 2. init database
	db, err := bootstrap.Database(conf, "0.0.0", false)
	if err != nil {
		t.Fatalf("No error expected: %v", err)
	}

	return db
}

func prepareData(db *sqlx.DB, w *workmodel.Work) (*workmodel.Work, error) {
	var err error

	w.ID, err = uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	ctx := context.Background()
	q := db.Rebind(`
		INSERT INTO works (id, start, stop) 
		VALUES (?, ?, ?);
	`)

	_, err = db.ExecContext(ctx, q, w.ID, w.Start, w.Stop)
	if err != nil {
		return nil, fmt.Errorf("failed to create data: %w", err)
	}

	ws := &workstore.Work{}

	q = db.Rebind(`SELECT id, start, stop, created_at, modified_at FROM works WHERE id = ?`)

	if err := db.GetContext(ctx, ws, q, w.ID); err != nil {
		return nil, fmt.Errorf("failed to retrieve created data: %w", err)
	}

	return workmapper.StoreToModel(ws), nil
}

func TestMapper_Load(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("long running test")
	}

	// 1. setup
	db := setup(t, "mapperLoad")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unable to close database connection: %v", err)
		}
	})

	mapper := workmapper.New(db)

	start, _ := time.Parse(time.RFC3339Nano, "2022-01-03T09:00:00.9168378+01:00")
	stop, _ := time.Parse(time.RFC3339Nano, "2022-01-03T13:00:00.9168378+01:00")

	// 2. test
	testCases := []struct {
		name        string
		id          uuid.UUID
		prepare     *workmodel.Work
		expected    *workmodel.Work
		expectedErr error
	}{
		{
			name: "success",
			prepare: &workmodel.Work{
				Start: start,
				Stop:  stop,
			},
			expected: &workmodel.Work{
				Start: start,
				Stop:  stop,
			},
		},
		{
			name:        "work not existing",
			id:          testingutils.UUIDParse(t, "9fcfda83-6749-4af4-a6eb-f7983218acf3"),
			expectedErr: workmapper.ErrNotFound,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var err error

			if testCase.prepare != nil {
				testCase.prepare, err = prepareData(db, testCase.prepare)
				if err != nil {
					t.Fatalf("failed to prepare data: %v", err)
				}

				testCase.id = testCase.prepare.ID
				testCase.expected.ID = testCase.prepare.ID
			}

			actual, err := mapper.Load(context.Background(), testCase.id)
			if !errors.Is(err, testCase.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", testCase.expectedErr, err)
			}

			assertWork(t, testCase.expected, actual)
		})
	}
}

func TestMapper_Save(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("long running test")
	}

	// 1. setup
	db := setup(t, "mapperSave")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unable to close database connection: %v", err)
		}
	})

	mapper := workmapper.New(db)

	start, _ := time.Parse(time.RFC3339Nano, "2022-01-03T09:00:00.9168378+01:00")
	stop, _ := time.Parse(time.RFC3339Nano, "2022-01-03T13:00:00.9168378+01:00")

	// 2. test
	testCases := []struct {
		name        string
		prepare     *workmodel.Work
		actual      *workmodel.Work
		expected    *workmodel.Work
		expectedErr error
		duplicate   bool
	}{
		{
			name:        "model is nil",
			expectedErr: workmapper.ErrNoData,
		},
		{
			name: "model has no ID",
			actual: &workmodel.Work{
				Start: start,
				Stop:  stop,
			},
			expected: &workmodel.Work{
				Start: start,
				Stop:  stop,
			},
		},
		{
			name: "model has ID",
			prepare: &workmodel.Work{
				Start: start,
				Stop:  stop,
			},
			actual: &workmodel.Work{
				Start: start,
				Stop:  stop,
			},
			expected: &workmodel.Work{
				Start: start,
				Stop:  stop,
			},
		},
		{
			name: "update not existing model",
			actual: &workmodel.Work{
				ID:    testingutils.UUIDParse(t, "efdaaa30-8a1c-47eb-9abf-173f2ae0e2a5"),
				Start: start,
				Stop:  stop,
			},
			expectedErr: workmapper.ErrSaveToDB,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var err error

			if testCase.prepare != nil {
				testCase.prepare, err = prepareData(db, testCase.prepare)
				if err != nil {
					t.Fatalf("failed to prepare data: %v", err)
				}

				if !testCase.duplicate {
					testCase.actual.ID = testCase.prepare.ID
				}
			}

			res, err := mapper.Save(context.Background(), testCase.actual)
			if !errors.Is(err, testCase.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", testCase.expectedErr, err)
			}

			if res != nil && testCase.expected != nil {
				testCase.expected.ID = res.ID
			}

			assertWork(t, testCase.expected, res)
		})
	}
}

func TestMapper_Delete(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("long running test")
	}

	// 1. setup
	db := setup(t, "mapperDelete")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unable to close database connection: %v", err)
		}
	})

	mapper := workmapper.New(db)

	start, _ := time.Parse(time.RFC3339Nano, "2022-01-03T09:00:00.9168378+01:00")
	stop, _ := time.Parse(time.RFC3339Nano, "2022-01-03T13:00:00.9168378+01:00")

	// 2. test
	testCases := []struct {
		name        string
		id          uuid.UUID
		prepare     *workmodel.Work
		expectedErr error
	}{
		{
			name: "success",
			prepare: &workmodel.Work{
				Start: start,
				Stop:  stop,
			},
		},
		{
			name: "work not existing",
			id:   testingutils.UUIDParse(t, "dda1dff5-7324-44d0-b04f-3ff25b80d896"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var err error

			if testCase.prepare != nil {
				testCase.prepare, err = prepareData(db, testCase.prepare)
				if err != nil {
					t.Fatalf("failed to prepare data: %v", err)
				}

				testCase.id = testCase.prepare.ID
			}

			err = mapper.Delete(context.Background(), testCase.id)
			if !errors.Is(err, testCase.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", testCase.expectedErr, err)

				return
			}

			if testCase.expectedErr == nil {
				_, err = mapper.Load(context.Background(), testCase.id)
				if !errors.Is(err, workmapper.ErrNotFound) {
					t.Errorf("expected that work was deleted but got error '%v'", err)
				}
			}
		})
	}
}

func TestMapper_StoreToModel(t *testing.T) {
	t.Parallel()

	start, _ := time.Parse(time.RFC3339Nano, "2022-01-03T09:00:00.9168378+01:00")
	stop, _ := time.Parse(time.RFC3339Nano, "2022-01-03T13:00:00.9168378+01:00")

	testCases := []struct {
		name     string
		actual   *workstore.Work
		expected *workmodel.Work
	}{
		{
			name:     "store is nil",
			expected: &workmodel.Work{},
		},
		{
			name: "store has all attributes set",
			actual: &workstore.Work{
				ID:    testingutils.UUIDParse(t, "5feb864a-6d39-4e0e-9bde-d9464cfcfba7"),
				Start: start,
				Stop:  stop,
			},
			expected: &workmodel.Work{
				ID:    testingutils.UUIDParse(t, "5feb864a-6d39-4e0e-9bde-d9464cfcfba7"),
				Start: start,
				Stop:  stop,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			u := workmapper.StoreToModel(testCase.actual)

			assertWork(t, testCase.expected, u)
		})
	}
}

func assertWork(t *testing.T, expected, actual *workmodel.Work) {
	t.Helper()

	if expected == nil && actual == nil {
		return
	}

	if expected != nil && actual == nil || expected == nil && actual != nil {
		t.Errorf("expected work '%v' but got '%v'", expected, actual)

		return
	}

	if expected.ID != actual.ID {
		t.Errorf("expected ID %s but got %s", expected.ID, actual.ID)
	}

	if !expected.Start.Equal(actual.Start) {
		t.Errorf("expected Start %s but got %s", expected.Start, actual.Start)
	}

	if !expected.Stop.Equal(actual.Stop) {
		t.Errorf("expected Stop %s but got %s", expected.Stop, actual.Stop)
	}

	if expected.CreatedAt != actual.CreatedAt && actual.CreatedAt.IsZero() {
		t.Error("created at should be greater than the zero date")
	}

	if expected.ModifiedAt != actual.ModifiedAt && actual.ModifiedAt.IsZero() {
		t.Error("created at should be greater than the zero date")
	}
}
