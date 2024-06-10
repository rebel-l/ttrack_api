package publicholidaymapper_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/go-utils/osutils"
	"github.com/rebel-l/go-utils/testingutils"
	"github.com/rebel-l/ttrack_api/bootstrap"
	"github.com/rebel-l/ttrack_api/config"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymapper"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
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

func prepareData(db *sqlx.DB, p *publicholidaymodel.PublicHoliday) (*publicholidaymodel.PublicHoliday, error) {
	var err error

	p.ID, err = uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	ctx := context.Background()
	q := db.Rebind(`
		INSERT INTO publicholidays (id, day, name, halfday) 
		VALUES (?, ?, ?, ?);
	`)

	_, err = db.ExecContext(ctx, q, p.ID, p.Day, p.Name, p.HalfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to create data: %w", err)
	}

	ps := &publicholidaystore.PublicHoliday{}

	q = db.Rebind(`SELECT id, day, name, halfday, created_at, modified_at FROM publicholidays WHERE id = ?`)

	if err := db.GetContext(ctx, ps, q, p.ID); err != nil {
		return nil, fmt.Errorf("failed to retrieve created data: %w", err)
	}

	return publicholidaymapper.StoreToModel(ps), nil
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

	mapper := publicholidaymapper.New(db)

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	// 2. test
	testCases := []struct {
		name        string
		id          uuid.UUID
		prepare     *publicholidaymodel.PublicHoliday
		expected    *publicholidaymodel.PublicHoliday
		expectedErr error
	}{
		{
			name: "success",
			prepare: &publicholidaymodel.PublicHoliday{
				Day:     now,
				Name:    "0IYnTmTIBsLOyNyN3RpvCyYwio7ikC4wsrUNHFQZQ9Lr4OIawI9x0KF7PbOOwIEZBx7QeErJijuGmaHQhh9XEH3rM13PxdMnnG3eN3SnRQbn3WOxOwb0uw2v14T8wT4QQAjNFTySIK2nCqaTloyPq9hcQiej7pxGF3XjJrhGayqXzNEQAGNFc2CnXEf900J45i21NiwPdRy8biAmv4zvOaFnTb9oZqApozoBdS4v",
				HalfDay: true,
			},
			expected: &publicholidaymodel.PublicHoliday{
				Day:     now,
				Name:    "0IYnTmTIBsLOyNyN3RpvCyYwio7ikC4wsrUNHFQZQ9Lr4OIawI9x0KF7PbOOwIEZBx7QeErJijuGmaHQhh9XEH3rM13PxdMnnG3eN3SnRQbn3WOxOwb0uw2v14T8wT4QQAjNFTySIK2nCqaTloyPq9hcQiej7pxGF3XjJrhGayqXzNEQAGNFc2CnXEf900J45i21NiwPdRy8biAmv4zvOaFnTb9oZqApozoBdS4v",
				HalfDay: true,
			},
		},
		{
			name:        "publicholiday not existing",
			id:          testingutils.UUIDParse(t, "31b0cef3-0e70-4932-8870-04a801861ffe"),
			expectedErr: publicholidaymapper.ErrNotFound,
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

			assertPublicHoliday(t, testCase.expected, actual)
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

	mapper := publicholidaymapper.New(db)

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	// 2. test
	testCases := []struct {
		name        string
		prepare     *publicholidaymodel.PublicHoliday
		actual      *publicholidaymodel.PublicHoliday
		expected    *publicholidaymodel.PublicHoliday
		expectedErr error
		duplicate   bool
	}{
		{
			name:        "model is nil",
			expectedErr: publicholidaymapper.ErrNoData,
		},
		{
			name: "model has no ID",
			actual: &publicholidaymodel.PublicHoliday{
				Day:     now,
				Name:    "o8dAhk0GIPNIxThxQ8nHyBAwomKYC57lyP1OO8y9PaRP8OUv5vi",
				HalfDay: true,
			},
			expected: &publicholidaymodel.PublicHoliday{
				Day:     now,
				Name:    "o8dAhk0GIPNIxThxQ8nHyBAwomKYC57lyP1OO8y9PaRP8OUv5vi",
				HalfDay: true,
			},
		},
		{
			name: "model has ID",
			prepare: &publicholidaymodel.PublicHoliday{
				Day:     now,
				Name:    "80WuukKQxbasVEL8wu7VJBQ8Ok91NnLntpjpdNuuReephgxFpkMY23aV1vB1iac0tlIwbbhWEor1NhthphLx6bqKbT2PeCiYc8FjLpyPXBoDhfbnWA2Eey9IoY4CNfckQVbbdUmU5lSbRoD9cE4cL0YtYGSm",
				HalfDay: true,
			},
			actual: &publicholidaymodel.PublicHoliday{
				Day:     now.Add(24 * time.Hour),
				Name:    "80WuukKQxbasVEL8wu7VJBQ8Ok91NnLntpjpdNuuReephgxFpkMY23aV1vB1iac0tlIwbbhWEor1NhthphLx6bqKbT2PeCiYc8FjLpyPXBoDhfbnWA2Eey9IoY4CNfckQVbbdUmU5lSbRoD9cE4cL0YtYGSm",
				HalfDay: true,
			},
			expected: &publicholidaymodel.PublicHoliday{
				Day:     now.Add(24 * time.Hour),
				Name:    "80WuukKQxbasVEL8wu7VJBQ8Ok91NnLntpjpdNuuReephgxFpkMY23aV1vB1iac0tlIwbbhWEor1NhthphLx6bqKbT2PeCiYc8FjLpyPXBoDhfbnWA2Eey9IoY4CNfckQVbbdUmU5lSbRoD9cE4cL0YtYGSm",
				HalfDay: true,
			},
		},
		{
			name: "update not existing model",
			actual: &publicholidaymodel.PublicHoliday{
				ID:      testingutils.UUIDParse(t, "63f3dc0b-54bf-4c5d-b577-5c20e047bd10"),
				Day:     now,
				Name:    "r8VzAcQEBwb7wxs0To47hUggccbjtKoyIb89nAYOj7CMDr7J6BBF99KdjosCUV0YSzqkQl7eYt7hZoJNAUfX1hOV3v1zbpbqvIqGj5q6XGmx25vzGdllXkEcswtsJkXvulLVOly0U2odyl9LKVoYeFhGL5IP3zm0J6htYL6LYKZwNnn",
				HalfDay: true,
			},
			expectedErr: publicholidaymapper.ErrSaveToDB,
		}}

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

			assertPublicHoliday(t, testCase.expected, res)
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

	mapper := publicholidaymapper.New(db)

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	// 2. test
	testCases := []struct {
		name        string
		id          uuid.UUID
		prepare     *publicholidaymodel.PublicHoliday
		expectedErr error
	}{
		{
			name: "success",
			prepare: &publicholidaymodel.PublicHoliday{
				Day:     now,
				Name:    "jDIV0xeStA3ALxamVibpKCZjiXw3esOHZDomI0ZRNNuw6PHetts2ZSCe0fKD0W0qNS6JQFZD1tu3hwKokXQqrGoqrxr1af0xddijjB",
				HalfDay: true,
			},
		},
		{
			name: "publicholiday not existing",
			id:   testingutils.UUIDParse(t, "f2b3fa9b-4d63-49a6-9dc9-9a4fd3d10fe6"),
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
				if !errors.Is(err, publicholidaymapper.ErrNotFound) {
					t.Errorf("expected that publicholiday was deleted but got error '%v'", err)
				}
			}
		})
	}
}

func TestMapper_StoreToModel(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		actual   *publicholidaystore.PublicHoliday
		expected *publicholidaymodel.PublicHoliday
	}{
		{
			name:     "store is nil",
			expected: &publicholidaymodel.PublicHoliday{},
		},
		{
			name: "store has all attributes set",
			actual: &publicholidaystore.PublicHoliday{
				ID:      testingutils.UUIDParse(t, "8435b797-6999-4a23-b2cc-256108e88a6c"),
				Day:     now,
				Name:    "u1ot5HbB1AW1WfcB9X77NtXXnghfsbuJMKR8IGrdMGkynQTNxFoSnjKmkcG6oZlSEZVNCfLs106If9pV21RZlx1nrTZDlhwRBDNC0dPkaLRrPlHllS1z2T3GzmqTCzjd7leGIQ7LrAEAQh4xQeN58PEr8GHxz64h4jgy7AND",
				HalfDay: true,
			},
			expected: &publicholidaymodel.PublicHoliday{
				ID:      testingutils.UUIDParse(t, "8435b797-6999-4a23-b2cc-256108e88a6c"),
				Day:     now,
				Name:    "u1ot5HbB1AW1WfcB9X77NtXXnghfsbuJMKR8IGrdMGkynQTNxFoSnjKmkcG6oZlSEZVNCfLs106If9pV21RZlx1nrTZDlhwRBDNC0dPkaLRrPlHllS1z2T3GzmqTCzjd7leGIQ7LrAEAQh4xQeN58PEr8GHxz64h4jgy7AND",
				HalfDay: true,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			u := publicholidaymapper.StoreToModel(testCase.actual)

			assertPublicHoliday(t, testCase.expected, u)
		})
	}
}

func assertPublicHoliday(t *testing.T, expected, actual *publicholidaymodel.PublicHoliday) {
	t.Helper()

	if expected == nil && actual == nil {
		return
	}

	if expected != nil && actual == nil || expected == nil && actual != nil {
		t.Errorf("expected publicholiday '%v' but got '%v'", expected, actual)

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

	if expected.CreatedAt != actual.CreatedAt && actual.CreatedAt.IsZero() {
		t.Error("created at should be greater than the zero date")
	}

	if expected.ModifiedAt != actual.ModifiedAt && actual.ModifiedAt.IsZero() {
		t.Error("created at should be greater than the zero date")
	}
}
