package timelogstore_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rebel-l/go-utils/osutils"
	"github.com/rebel-l/go-utils/testingutils"
	"github.com/rebel-l/go-utils/uuidutils"
	"github.com/rebel-l/ttrack_api/bootstrap"
	"github.com/rebel-l/ttrack_api/config"
	"github.com/rebel-l/ttrack_api/timelog/timelogstore"
)

func setup(t *testing.T, name string) *sqlx.DB {
	t.Helper()

	// 0. init path
	storagePath := filepath.Join(".", "..", "..", "storage", "test_timelog", name)

	// nolint: godox
	// TODO: change that it works with other dialects like postgres
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

func TestTimelog_Create(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("long running test")
	}

	// 1. setup
	db := setup(t, "storeCreate")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unable to close database connection: %v", err)
		}
	})

	// 2. test
	testCases := []struct {
		name        string
		actual      *timelogstore.Timelog
		expected    *timelogstore.Timelog
		expectedErr error
	}{
		{
			name:        "timelog is nil",
			expectedErr: timelogstore.ErrDataMissing,
		},
		{
			name: "timelog has stop only",
			actual: &timelogstore.Timelog{
				Stop: testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
			},
			expectedErr: timelogstore.ErrDataMissing,
		},
		{
			name: "timelog has id",
			actual: &timelogstore.Timelog{
				ID:       testingutils.UUIDParse(t, "befaea68-6ae2-435f-a47c-0e247c9976b4"),
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
				Reason:   "a6N2b8bwM9Lbgey",
				Location: "Lg036olw0Q",
			},
			expectedErr: timelogstore.ErrIDIsSet,
		},
		{
			name: "timelog has all fields set",
			actual: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
				Reason:   "o55jR",
				Location: "pNDJG",
			},
			expected: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
				Reason:   "o55jR",
				Location: "pNDJG",
			},
		},
		{
			name: "timelog has only mandatory fields set",
			actual: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
				Reason:   "bGXFPE3aKo",
				Location: "UL1nJANso3zIbU7jA",
			},
			expected: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5171183+01:00"),
				Reason:   "bGXFPE3aKo",
				Location: "UL1nJANso3zIbU7jA",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.actual.Create(context.Background(), db)
			testingutils.ErrorsCheck(t, testCase.expectedErr, err)

			if testCase.expectedErr == nil {
				testCase.expected.ID = testCase.actual.ID
				assertTimelog(t, testCase.expected, testCase.actual)
			}
		})
	}
}

func TestTimelog_Read(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("long running test")
	}

	// 1. setup
	db := setup(t, "storeRead")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unable to close database connection: %v", err)
		}
	})

	// 2. test
	testCases := []struct {
		name        string
		prepare     *timelogstore.Timelog
		expected    *timelogstore.Timelog
		expectedErr error
	}{
		{
			name:        "timelog is nil",
			expectedErr: timelogstore.ErrIDMissing,
		},
		{
			name:        "ID not set",
			expectedErr: timelogstore.ErrIDMissing,
		},
		{
			name: "success",
			prepare: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5176385+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5176385+01:00"),
				Reason:   "u90uSdRg",
				Location: "hQEl5UELFQe9hXF1vPxc",
			},
			expected: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5176385+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5176385+01:00"),
				Reason:   "u90uSdRg",
				Location: "hQEl5UELFQe9hXF1vPxc",
			},
		},
		{
			name: "not existing",
			prepare: &timelogstore.Timelog{
				ID: testingutils.UUIDParse(t, "5c050608-e2e4-4410-b5d8-cefa02a8eff0"),
			},
			expectedErr: sql.ErrNoRows,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var id uuid.UUID
			if testCase.prepare != nil {
				if testCase.prepare.IsValid() {
					err := testCase.prepare.Create(context.Background(), db)
					if err != nil {
						t.Errorf("preparation failed: %v", err)

						return
					}
				}
				id = testCase.prepare.ID
			}

			actual := &timelogstore.Timelog{ID: id}
			err := actual.Read(context.Background(), db)
			testingutils.ErrorsCheck(t, testCase.expectedErr, err)

			if testCase.expectedErr == nil {
				testCase.expected.ID = actual.ID
				assertTimelog(t, testCase.expected, actual)
			}
		})
	}
}

func assertTimelog(t *testing.T, expected, actual *timelogstore.Timelog) {
	t.Helper()

	if expected == nil && actual == nil {
		return
	}

	if expected != nil && actual == nil || expected == nil && actual != nil {
		t.Errorf("expected '%v' but got '%v'", expected, actual)

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

	if actual.CreatedAt.IsZero() {
		t.Error("created at should be greater than the zero date")
	}

	if actual.ModifiedAt.IsZero() {
		t.Error("modified at should be greater than the zero date")
	}
}

func TestTimelog_Update(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("long running test")
	}

	// 1. setup
	db := setup(t, "storeUpdate")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unable to close database connection: %v", err)
		}
	})

	// 2. test
	testCases := []struct {
		name        string
		prepare     *timelogstore.Timelog
		actual      *timelogstore.Timelog
		expected    *timelogstore.Timelog
		expectedErr error
	}{
		{
			name:        "timelog is nil",
			expectedErr: timelogstore.ErrDataMissing,
		},
		{
			name: "timelog has stop only",
			actual: &timelogstore.Timelog{
				Stop: testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
			},
			expectedErr: timelogstore.ErrDataMissing,
		},
		{
			name: "timelog has no id",
			actual: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Reason:   "AUoO5KDOGsEi6FCN3tK7xzm",
				Location: "I5FzQfgkuDVHpigHqN3M",
			},
			expectedErr: timelogstore.ErrIDMissing,
		},
		{
			name: "not existing",
			actual: &timelogstore.Timelog{
				ID:       testingutils.UUIDParse(t, "4cd348a5-0e8f-4986-b800-540ef9a0775a"),
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Reason:   "HZlgkz3F1FWKfUIh",
				Location: "EINneuX",
			},
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "timelog has all fields set",
			actual: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Reason:   "jzaCnKKSRKmtLD3",
				Location: "0rn6QkWPCIc",
			},
			prepare: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Reason:   "7ugUH7xdgUo9Wj68bo6ayKTQymKEy",
				Location: "ZOn6KufVUxaotfkh",
			},
			expected: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Reason:   "jzaCnKKSRKmtLD3",
				Location: "0rn6QkWPCIc",
			},
		},
		{
			name: "timelog has only mandatory fields set",
			actual: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Reason:   "4k6BbZT2EhAuZ",
				Location: "i8jcjpO3Glq52UDlrc",
			},
			prepare: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Reason:   "FUj0KKBai9pUo8b3Sicl5",
				Location: "uIF3cL5SYIWlB87jDd",
			},
			expected: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5186965+01:00"),
				Reason:   "4k6BbZT2EhAuZ",
				Location: "i8jcjpO3Glq52UDlrc",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.prepare != nil {
				_ = testCase.prepare.Create(context.Background(), db)
				time.Sleep(1 * time.Second)
				testCase.actual.ID = testCase.prepare.ID
			}

			err := testCase.actual.Update(context.Background(), db)
			testingutils.ErrorsCheck(t, testCase.expectedErr, err)

			if testCase.expectedErr == nil {
				testCase.expected.ID = testCase.actual.ID
				assertTimelog(t, testCase.expected, testCase.actual)
			}

			if testCase.prepare != nil && testCase.actual != nil {
				if testCase.prepare.CreatedAt != testCase.actual.CreatedAt {
					t.Errorf(
						"expected created at '%s' but got '%s'",
						testCase.prepare.CreatedAt.String(),
						testCase.actual.CreatedAt.String(),
					)
				}

				if testCase.prepare.ModifiedAt.After(testCase.actual.ModifiedAt) {
					t.Errorf(
						"expected modified at '%s' to be before but got '%s'",
						testCase.prepare.ModifiedAt.String(),
						testCase.actual.ModifiedAt.String(),
					)
				}
			}
		})
	}
}

func TestTimelog_Delete(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("long running test")
	}

	// 1. setup
	db := setup(t, "storeDelete")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unable to close database connection: %v", err)
		}
	})

	// 2. test
	testCases := []struct {
		name        string
		prepare     *timelogstore.Timelog
		expectedErr error
	}{
		{
			name:        "timelog is nil",
			expectedErr: timelogstore.ErrIDMissing,
		},
		{
			name:        "timelog has no ID",
			expectedErr: timelogstore.ErrIDMissing,
		},
		{
			name: "success",
			prepare: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
				Reason:   "YoElEQ",
				Location: "9gSjvLOuLH35t9eqNmN",
			},
		},
		{
			name: "not existing",
			prepare: &timelogstore.Timelog{
				ID: testingutils.UUIDParse(t, "34dbbd09-af9e-4e33-9f12-42a4a9b24315"),
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var id uuid.UUID
			if testCase.prepare != nil {
				if testCase.prepare.IsValid() {
					err := testCase.prepare.Create(context.Background(), db)
					if err != nil {
						t.Errorf("preparation failed: %v", err)

						return
					}
				}
				id = testCase.prepare.ID
			}

			actual := &timelogstore.Timelog{ID: id}
			err := actual.Delete(context.Background(), db)
			testingutils.ErrorsCheck(t, testCase.expectedErr, err)

			if !uuidutils.IsEmpty(id) {
				err := actual.Read(context.Background(), db)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Errorf("expected error '%v' after deletion but got '%v'", sql.ErrNoRows, err)
				}
			}
		})
	}
}

func TestTimelog_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		actual   *timelogstore.Timelog
		expected bool
	}{
		{
			name:     "timelog is nil",
			expected: false,
		},
		{
			name: "timelog has id only",
			actual: &timelogstore.Timelog{
				ID: testingutils.UUIDParse(t, "a045fa97-1d5d-4569-aa9d-0a3f0ca5a230"),
			},
			expected: false,
		},
		{
			name: "timelog has start only",
			actual: &timelogstore.Timelog{
				Start: testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
			},
			expected: false,
		},
		{
			name: "timelog has stop only",
			actual: &timelogstore.Timelog{
				Stop: testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
			},
			expected: false,
		},
		{
			name: "timelog has reason only",
			actual: &timelogstore.Timelog{
				Reason: "qJgeORL8sru2EoZxv9cqxkDWhVh",
			},
			expected: false,
		},
		{
			name: "timelog has location only",
			actual: &timelogstore.Timelog{
				Location: "GZ960aA",
			},
			expected: false,
		},
		{
			name: "mandatory fields only",
			actual: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
				Reason:   "rIfOyu9S5kSAmHJUkdKfjohIPfH",
				Location: "w70uz",
			},
			expected: true,
		},
		{
			name: "mandatory fields with id",
			actual: &timelogstore.Timelog{
				ID:       testingutils.UUIDParse(t, "469c4110-e632-4b1b-b561-d90fd639fcf6"),
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
				Reason:   "HCh9wo5vPUplNkfOv5ztYQz970Qsns",
				Location: "XpVSsN4mjpBL9H",
			},
			expected: true,
		},
		{
			name: "all fields",
			actual: &timelogstore.Timelog{
				ID:       testingutils.UUIDParse(t, "2cfd9425-6ed4-4a06-9eab-eed2c818521f"),
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
				Reason:   "ilKCRpvFg",
				Location: "73NcOR7TOYhqOJEkMfE",
			},
			expected: true,
		},
		{
			name: "all fields without id",
			actual: &timelogstore.Timelog{
				Start:    testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
				Stop:     testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5197509+01:00"),
				Reason:   "bUbRNYpgHt",
				Location: "CZ7h874jfupnvww",
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			res := testCase.actual.IsValid()
			if testCase.expected != res {
				t.Errorf("expected %t but got %t", testCase.expected, res)
			}
		})
	}
}
