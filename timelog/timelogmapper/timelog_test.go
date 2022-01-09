package timelogmapper_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rebel-l/go-utils/osutils"
	"github.com/rebel-l/go-utils/testingutils"
	"github.com/rebel-l/ttrack_api/bootstrap"
	"github.com/rebel-l/ttrack_api/config"
	"github.com/rebel-l/ttrack_api/timelog/timelogmapper"
	"github.com/rebel-l/ttrack_api/timelog/timelogmodel"
	"github.com/rebel-l/ttrack_api/timelog/timelogstore"
)

func setup(t *testing.T, name string) *sqlx.DB {
	t.Helper()

	// 0. init path
	storagePath := filepath.Join(".", "..", "..", "storage", "test_timelog", name)
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

func prepareData(db *sqlx.DB, t *timelogmodel.Timelog) (*timelogmodel.Timelog, error) {
	var err error

	t.ID, err = uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	ctx := context.Background()
	q := db.Rebind(`
		INSERT INTO timelogs (id, start, stop, reason, location) 
		VALUES (?, ?, ?, ?, ?);
	`)

	_, err = db.ExecContext(ctx, q, t.ID, t.Start, t.Stop, t.Reason, t.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to create data: %w", err)
	}

	ts := &timelogstore.Timelog{}

	q = db.Rebind(`SELECT id, start, stop, reason, location, created_at, modified_at FROM timelogs WHERE id = ?`)

	if err := db.GetContext(ctx, ts, q, t.ID); err != nil {
		return nil, fmt.Errorf("failed to retrieve created data: %w", err)
	}

	return timelogmapper.StoreToModel(ts), nil
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

	mapper := timelogmapper.New(db)

	// 2. test
	testCases := []struct {
		name        string
		id          uuid.UUID
		prepare     *timelogmodel.Timelog
		expected    *timelogmodel.Timelog
		expectedErr error
	}{
		{
			name: "success",
			prepare: &timelogmodel.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Reason:   "gJ5bvmhH7n69mYJnrrFdreh",
				Location: "gXlRC",
			},
			expected: &timelogmodel.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Reason:   "gJ5bvmhH7n69mYJnrrFdreh",
				Location: "gXlRC",
			},
		},
		{
			name:        "timelog not existing",
			id:          testingutils.UUIDParse(t, "990db276-6b6e-491a-a312-46e32f4adb71"),
			expectedErr: timelogmapper.ErrNotFound,
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

			assertTimelog(t, testCase.expected, actual)
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

	mapper := timelogmapper.New(db)

	// 2. test
	testCases := []struct {
		name        string
		prepare     *timelogmodel.Timelog
		actual      *timelogmodel.Timelog
		expected    *timelogmodel.Timelog
		expectedErr error
		duplicate   bool
	}{
		{
			name:        "model is nil",
			expectedErr: timelogmapper.ErrNoData,
		},
		{
			name: "model has no ID",
			actual: &timelogmodel.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Reason:   "gRPoU4fIqsE98Vg4OXV7ZBI3YpbW3V",
				Location: "mMGzBfuUMS0n7MI8",
			},
			expected: &timelogmodel.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Reason:   "gRPoU4fIqsE98Vg4OXV7ZBI3YpbW3V",
				Location: "mMGzBfuUMS0n7MI8",
			},
		},
		{
			name: "model has ID",
			prepare: &timelogmodel.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Reason:   "NJlSJ5t",
				Location: "DnTTVROZ2dneNsNxE",
			},
			actual: &timelogmodel.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Reason:   "NJlSJ5t",
				Location: "DnTTVROZ2dneNsNxE",
			},
			expected: &timelogmodel.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Reason:   "NJlSJ5t",
				Location: "DnTTVROZ2dneNsNxE",
			},
		},
		{
			name: "update not existing model",
			actual: &timelogmodel.Timelog{
				ID:       testingutils.UUIDParse(t, "7bef48b3-ec44-4bfa-9d76-52351a4a08ff"),
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5244444+01:00"),
				Reason:   "hlVhuo8v7Pre897zTYH3cCyt",
				Location: "Ks5vh2SnPX0",
			},
			expectedErr: timelogmapper.ErrSaveToDB,
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

			assertTimelog(t, testCase.expected, res)
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

	mapper := timelogmapper.New(db)

	// 2. test
	testCases := []struct {
		name        string
		id          uuid.UUID
		prepare     *timelogmodel.Timelog
		expectedErr error
	}{
		{
			name: "success",
			prepare: &timelogmodel.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.524966+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.524966+01:00"),
				Reason:   "mgt6EDzh5dIOUWJbZioto6i",
				Location: "3nPDy54NTbO3",
			},
		},
		{
			name: "timelog not existing",
			id:   testingutils.UUIDParse(t, "9468aa5c-b72f-4be5-a637-08c9cf21da6f"),
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
				if !errors.Is(err, timelogmapper.ErrNotFound) {
					t.Errorf("expected that timelog was deleted but got error '%v'", err)
				}
			}
		})
	}
}

func TestMapper_StoreToModel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		actual   *timelogstore.Timelog
		expected *timelogmodel.Timelog
	}{
		{
			name:     "store is nil",
			expected: &timelogmodel.Timelog{},
		},
		{
			name: "store has all attributes set",
			actual: &timelogstore.Timelog{
				ID:       testingutils.UUIDParse(t, "6fd7a340-e733-41fc-800f-24da25f20a81"),
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.524966+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.524966+01:00"),
				Reason:   "JcUKO5AFnBrqHX0DfFeEuKvXLmS",
				Location: "Y48thALSKldPWtv8",
			},
			expected: &timelogmodel.Timelog{
				ID:       testingutils.UUIDParse(t, "6fd7a340-e733-41fc-800f-24da25f20a81"),
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.524966+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.524966+01:00"),
				Reason:   "JcUKO5AFnBrqHX0DfFeEuKvXLmS",
				Location: "Y48thALSKldPWtv8",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			u := timelogmapper.StoreToModel(testCase.actual)

			assertTimelog(t, testCase.expected, u)
		})
	}
}

func assertTimelog(t *testing.T, expected, actual *timelogmodel.Timelog) {
	t.Helper()

	if expected == nil && actual == nil {
		return
	}

	if expected != nil && actual == nil || expected == nil && actual != nil {
		t.Errorf("expected timelog '%v' but got '%v'", expected, actual)

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

	if expected.Reason != actual.Reason {
		t.Errorf("expected Reason %s but got %s", expected.Reason, actual.Reason)
	}

	if expected.Location != actual.Location {
		t.Errorf("expected Location %s but got %s", expected.Location, actual.Location)
	}

	if expected.CreatedAt != actual.CreatedAt && actual.CreatedAt.IsZero() {
		t.Error("created at should be greater than the zero date")
	}

	if expected.ModifiedAt != actual.ModifiedAt && actual.ModifiedAt.IsZero() {
		t.Error("created at should be greater than the zero date")
	}
}
