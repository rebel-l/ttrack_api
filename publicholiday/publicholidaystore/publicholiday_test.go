package publicholidaystore_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/go-utils/osutils"
	"github.com/rebel-l/go-utils/testingutils"
	"github.com/rebel-l/go-utils/uuidutils"
	"github.com/rebel-l/ttrack_api/bootstrap"
	"github.com/rebel-l/ttrack_api/config"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaystore"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setup(t *testing.T, name string) *sqlx.DB {
	t.Helper()

	// 0. init path
	storagePath := filepath.Join(".", "..", "..", "storage", "test_publicholiday", name)

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

func TestPublicHoliday_Create(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("long running test")
	}

	// 1. setup
	db := setup(t, "storeCreate")

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("unable to close database connection: %v", err)
		}
	})

	// 2. test
	testCases := []struct {
		name        string
		actual      *publicholidaystore.PublicHoliday
		expected    *publicholidaystore.PublicHoliday
		expectedErr error
	}{
		{
			name:        "publicholiday is nil",
			expectedErr: publicholidaystore.ErrDataMissing,
		},
		{
			name: "publicholiday has id",
			actual: &publicholidaystore.PublicHoliday{
				ID:      testingutils.UUIDParse(t, "2d8dc0c0-927f-4e74-adfd-6d803421347a"),
				Day:     now,
				Name:    "NSiDl1JyDEC6f7bzAg53csEtj1ZPEyc9ArTmnevJueK7hXnshI0eb6En4znyqtjlcATu7EiDpl6UL3A816mtDDyaNV0LzWDMg4xFJeBHVpNR96UYmRFylw6gYkboFQAo7wSQX13vf9w8lUIgeyB0FvWw1MDwpD7Q5afXgxawvlUAt0kNR6dvg7kF84TFel0pbnM4OS3ITyfZLLB6VlGGRr",
				HalfDay: true,
			},
			expectedErr: publicholidaystore.ErrIDIsSet,
		},
		{
			name: "publicholiday has all fields set",
			actual: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "SHhpY72xVwW2FMCKSC5X7v6DzSAGoIL3qwNY",
				HalfDay: true,
			},
			expected: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "SHhpY72xVwW2FMCKSC5X7v6DzSAGoIL3qwNY",
				HalfDay: true,
			},
		},
		{
			name: "publicholiday has only mandatory fields set",
			actual: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "iXmVQh4iYmmvHzyy7fydphL5PmGQwxK",
				HalfDay: true,
			},
			expected: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "iXmVQh4iYmmvHzyy7fydphL5PmGQwxK",
				HalfDay: true,
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
				assertPublicHoliday(t, testCase.expected, testCase.actual)
			}
		})
	}
}

func TestPublicHoliday_Read(t *testing.T) {
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

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	// 2. test
	testCases := []struct {
		name        string
		prepare     *publicholidaystore.PublicHoliday
		expected    *publicholidaystore.PublicHoliday
		expectedErr error
	}{
		{
			name:        "publicholiday is nil",
			expectedErr: publicholidaystore.ErrIDMissing,
		},
		{
			name:        "ID not set",
			expectedErr: publicholidaystore.ErrIDMissing,
		},
		{
			name: "success",
			prepare: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "NxN2kMEwq0HeVa6HHoNNnvFjnsUFzhTHRX5dIr6e1bTwMHVzdOVYWjEYiLdPxYe7euupdWAhE3V4nzD91WasOBPk7mE2xPWwHi8Jk75Me19WPKJrY9NivHp0RaEupSaHCuzn6bKi",
				HalfDay: true,
			},
			expected: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "NxN2kMEwq0HeVa6HHoNNnvFjnsUFzhTHRX5dIr6e1bTwMHVzdOVYWjEYiLdPxYe7euupdWAhE3V4nzD91WasOBPk7mE2xPWwHi8Jk75Me19WPKJrY9NivHp0RaEupSaHCuzn6bKi",
				HalfDay: true,
			},
		},
		{
			name: "not existing",
			prepare: &publicholidaystore.PublicHoliday{
				ID: testingutils.UUIDParse(t, "411484c9-6143-4d77-8cb5-d81af3937673"),
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

			actual := &publicholidaystore.PublicHoliday{ID: id}
			err := actual.Read(context.Background(), db)
			testingutils.ErrorsCheck(t, testCase.expectedErr, err)

			if testCase.expectedErr == nil {
				testCase.expected.ID = actual.ID
				assertPublicHoliday(t, testCase.expected, actual)
			}
		})
	}
}

func assertPublicHoliday(t *testing.T, expected, actual *publicholidaystore.PublicHoliday) {
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

	if !expected.Day.Equal(actual.Day) {
		t.Errorf("expected Day %q but got %q", expected.Day, actual.Day)
	}

	if expected.Name != actual.Name {
		t.Errorf("expected Name %s but got %s", expected.Name, actual.Name)
	}

	if expected.HalfDay != actual.HalfDay {
		t.Errorf("expected HalfDay %t but got %t", expected.HalfDay, actual.HalfDay)
	}

	if actual.CreatedAt.IsZero() {
		t.Error("created at should be greater than the zero date")
	}

	if actual.ModifiedAt.IsZero() {
		t.Error("modified at should be greater than the zero date")
	}
}

func TestPublicHoliday_Update(t *testing.T) {
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

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	// 2. test
	testCases := []struct {
		name        string
		prepare     *publicholidaystore.PublicHoliday
		actual      *publicholidaystore.PublicHoliday
		expected    *publicholidaystore.PublicHoliday
		expectedErr error
	}{
		{
			name:        "publicholiday is nil",
			expectedErr: publicholidaystore.ErrDataMissing,
		},
		{
			name: "publicholiday has no id",
			actual: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "MGTPn1zeHf7molyKR1OCqbVp3EXQxr3wn4uwKNQMyw2HTIsHD1F7Oz9P0FSjsb6RFPXTpSGiEiWj7piGebFna1ldYD03KhtzePbtAHvLtEY1Kk1oM9T3fenOYR2AqX8QoBpzfjtJn7FwL1KSyX4bxCqZZhgXnZtLbswlHIM5cjYphNRoZG6oI7i67tmPEC7xgW5dw4CQK",
				HalfDay: true,
			},
			expectedErr: publicholidaystore.ErrIDMissing,
		},
		{
			name: "not existing",
			actual: &publicholidaystore.PublicHoliday{
				ID:      testingutils.UUIDParse(t, "7051ce69-9979-4a26-94bf-bc621f71568a"),
				Day:     now,
				Name:    "kICyMuQlHelHORC8PyGDXbfcuHQxhi8hOx55LnNLscfBbOjPxygBQUZ",
				HalfDay: true,
			},
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "publicholiday has all fields set",
			actual: &publicholidaystore.PublicHoliday{
				Day:     now.Add(24 * time.Hour),
				Name:    "RkghL3fvOugDXccRCgjvlvNuesN2YaAAiwI1e77jdSN1IvLVaVnYBYM9YXaQghNWArZEbjPgl81M1JpEqDtsNTTtTaHxoXsqWbqbyHwE7x70mw7xMH4iMgNAE9Tk2avm8PXJ7V5U4WDL",
				HalfDay: true,
			},
			prepare: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "cuRrx4qRsVySGtaariYLvkxN48Fn5tlTwdAaaTexMdtLlaE",
				HalfDay: true,
			},
			expected: &publicholidaystore.PublicHoliday{
				Day:     now.Add(24 * time.Hour),
				Name:    "RkghL3fvOugDXccRCgjvlvNuesN2YaAAiwI1e77jdSN1IvLVaVnYBYM9YXaQghNWArZEbjPgl81M1JpEqDtsNTTtTaHxoXsqWbqbyHwE7x70mw7xMH4iMgNAE9Tk2avm8PXJ7V5U4WDL",
				HalfDay: true,
			},
		},
		{
			name: "publicholiday has only mandatory fields set",
			actual: &publicholidaystore.PublicHoliday{
				Day:     now.Add(48 * time.Hour),
				Name:    "TezPamk3B3hDQem0ydTmbmDAWJ4Mqovn0ndhSP0W9wOvx1UBWbgexF69h8pW76l0gubYoZFWXVwJ25JtuAfObJ49NgZ8tYiUtxwuPTxlF3XEOpnVBENtIWUxgQ8DOAj",
				HalfDay: true,
			},
			prepare: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "C67LgXWkQbJlpIiLfFyDfD6cUnvk0qLTv7XM0clirOun7BI1u3TWGkrIOQns4zi3mCqq3uBgeM6A3l5GOK0sQ6oOVLjeI5zWf6WijwfuXQH6bcaAKfxIYaWbGIOnjEZ5v68YNMYllPegXXnk9oddWzjengOr5CIeVoulJKMKcOdcLav317286QORcsKWLtOfRWWD2p0bsQXUu6p2Ii1mxW431F3dBQtqy57FK35OgGPQOQMETse",
				HalfDay: true,
			},
			expected: &publicholidaystore.PublicHoliday{
				Day:     now.Add(48 * time.Hour),
				Name:    "TezPamk3B3hDQem0ydTmbmDAWJ4Mqovn0ndhSP0W9wOvx1UBWbgexF69h8pW76l0gubYoZFWXVwJ25JtuAfObJ49NgZ8tYiUtxwuPTxlF3XEOpnVBENtIWUxgQ8DOAj",
				HalfDay: true,
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
				assertPublicHoliday(t, testCase.expected, testCase.actual)
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

func TestPublicHoliday_Delete(t *testing.T) {
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

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	// 2. test
	testCases := []struct {
		name        string
		prepare     *publicholidaystore.PublicHoliday
		expectedErr error
	}{
		{
			name:        "publicholiday is nil",
			expectedErr: publicholidaystore.ErrIDMissing,
		},
		{
			name:        "publicholiday has no ID",
			expectedErr: publicholidaystore.ErrIDMissing,
		},
		{
			name: "success",
			prepare: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "0kNw8R0tTxLCW4yxO3DcEK3w7gSB6mHRuw1q7ysmxEcdP0OuqA5XkuFoKOMT2urk8CDsDLMjfPk4t4scuwWRaJ0XuMfs45479lboXdPKQJZgfx21Moqwlr727v2cAGhNmVH8IVIIBz5kP89b08Umlldoi1pb64RCdwA",
				HalfDay: true,
			},
		},
		{
			name: "not existing",
			prepare: &publicholidaystore.PublicHoliday{
				ID: testingutils.UUIDParse(t, "b42e6b6b-2ea6-46af-b8e7-273b78626dd4"),
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

			actual := &publicholidaystore.PublicHoliday{ID: id}
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

func TestPublicHoliday_IsValid(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		actual   *publicholidaystore.PublicHoliday
		expected bool
	}{
		{
			name:     "publicholiday is nil",
			expected: false,
		},
		{
			name: "publicholiday has id only",
			actual: &publicholidaystore.PublicHoliday{
				ID: testingutils.UUIDParse(t, "bc9289f0-60a2-4cfe-b8bd-305ef2ae6374"),
			},
			expected: false,
		},
		{
			name: "publicholiday has day only",
			actual: &publicholidaystore.PublicHoliday{
				Day: now,
			},
			expected: false,
		},
		{
			name: "publicholiday has name only",
			actual: &publicholidaystore.PublicHoliday{
				Name: "t2E832RB4eCazHz4ofYsbUjawk07QmMfObBCJL3sBDctpYTC8mCezqiOX87mauMaheJd2rDPe3UWPOuUg5F5UG4PbrYJWrR5GGI0ICfcmaDhEvSaOrimOXrxGdTePLlGJkVQssekdzmRQNTuHl4tX5xJVGjA4hBi858KDxcL3Nv0WNXw9LiPIW2dIX5HgFcrQmpx8zhkz2acOm7fue8Waay0JtS3FfptBqJ",
			},
			expected: false,
		},
		{
			name: "publicholiday has halfday only",
			actual: &publicholidaystore.PublicHoliday{
				HalfDay: true,
			},
			expected: false,
		},
		{
			name: "mandatory fields only",
			actual: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "Wp1vfWIIE1Lwg0zun5Ocf5PSZ3RlyB9bIEsQhsxIshDI6PFOIJ5lWgQYBApLYRkYhFUXT4MoXoaQ1HJ0gISWb2lSKbm8OMP3pBS4hSsefkYhKwD2VtZo18os6aLoL8BmrvZhL26Ot5GbCPRuVa",
				HalfDay: true,
			},
			expected: true,
		},
		{
			name: "mandatory fields with id",
			actual: &publicholidaystore.PublicHoliday{
				ID:      testingutils.UUIDParse(t, "da5d2f18-448d-4ef4-bf7a-13d6dc84cb39"),
				Day:     now,
				Name:    "uKLengFnJmFZ0S3hyreNv9NRP9kSc9QLXxEnURmboJYxG07HyUZ9uJpePkH4oJWzroZpXeI0hAuE0MKGgHE9KuCvign5zqVEnpSqWHZFlowm2FusNmKPHinsekmN568EeOohr81NDpZRKkNCT9fJ9i4tyMaHk77aqIgmvAOABN432vwwr2zbNTYnWOfjip1m8XauL1ahjQZHCqIRiXh08s1Lift971",
				HalfDay: true,
			},
			expected: true,
		},
		{
			name: "all fields",
			actual: &publicholidaystore.PublicHoliday{
				ID:      testingutils.UUIDParse(t, "da6ce09f-5801-412a-a412-f4932945f3eb"),
				Day:     now,
				Name:    "rICgelUkOtOwLM8J6XizrfIwZLKOLY44wlO7N32cFNPTYNptbHdUlorRFCfju1CertznomUS",
				HalfDay: true,
			},
			expected: true,
		},
		{
			name: "all fields without id",
			actual: &publicholidaystore.PublicHoliday{
				Day:     now,
				Name:    "kiE0rxzncioLMHhRFbGiX2aKSOMm3p6cGQ5S",
				HalfDay: true,
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
