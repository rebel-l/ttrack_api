package workstore_test

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
    "github.com/rebel-l/go-utils/osutils"
    "github.com/rebel-l/go-utils/testingutils"
    "github.com/rebel-l/go-utils/uuidutils"
    "github.com/rebel-l/ttrack_api/bootstrap"
    "github.com/rebel-l/ttrack_api/config"
    "github.com/rebel-l/ttrack_api/work/workstore"

    _ "github.com/mattn/go-sqlite3"
)

func setup(t *testing.T, name string) *sqlx.DB {
    t.Helper()

    // 0. init path
    storagePath := filepath.Join(".", "..", "..", "storage", "test_work", name)

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

func TestWork_Create(t *testing.T) {
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
        actual      *workstore.Work
        expected    *workstore.Work
        expectedErr error
    }{
        {
            name:        "work is nil",
            expectedErr: workstore.ErrDataMissing,
        },
        {
            name:        "work has id only",
            actual:      &workstore.Work{
        ID: testingutils.UUIDParse(t, "134a74ee-153f-48de-a319-c510643353d1"),
},
            expectedErr: workstore.ErrDataMissing,
        },
        {
            name:        "work has stop only",
            actual:      &workstore.Work{
        Stop: ,
},
            expectedErr: workstore.ErrDataMissing,
        },
        {
            name:        "work has id",
            actual:      &workstore.Work{
        ID: testingutils.UUIDParse(t, "f7a92808-7d39-4e35-bc91-4829d50ccb45"),
        Start: "UWiZCkv",
        Stop: ,
},
            expectedErr: workstore.ErrIDIsSet,
        },
        {
            name:        "work has all fields set",
            actual:      &workstore.Work{
        Start: "OPNJW39E8xzp8ZrVaPNZbeyWiJlE3hfKHPZmT",
        Stop: ,
},
            expected: &workstore.Work{
        Start: "OPNJW39E8xzp8ZrVaPNZbeyWiJlE3hfKHPZmT",
        Stop: ,
},
        },
        {
            name:        "work has only mandatory fields set",
            actual:      &workstore.Work{
        Start: "61cdXLSI",
},
            expected: &workstore.Work{
        Start: "61cdXLSI",
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
                assertWork(t, testCase.expected, testCase.actual)
            }
        })
    }
}

func TestWork_Read(t *testing.T) {
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
        prepare     *workstore.Work
        expected    *workstore.Work
        expectedErr error
    }{
        {
            name:        "work is nil",
            expectedErr: workstore.ErrIDMissing,
        },
        {
            name:        "ID not set",
            expectedErr: workstore.ErrIDMissing,
        },
        {
            name:        "success",
            prepare:      &workstore.Work{
        Start: "jg4nHRr2vth969KY2CBdYmOXJD",
        Stop: ,
},
            expected: &workstore.Work{
        Start: "jg4nHRr2vth969KY2CBdYmOXJD",
        Stop: ,
},
        },
        {
            name:        "not existing",
            prepare:      &workstore.Work{
        ID: testingutils.UUIDParse(t, "c0175af8-ae74-4568-9168-16b3e71841e4"),
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

            actual := &workstore.Work{ID: id}
            err := actual.Read(context.Background(), db)
            testingutils.ErrorsCheck(t, testCase.expectedErr, err)

            if testCase.expectedErr == nil {
                testCase.expected.ID = actual.ID
                assertWork(t, testCase.expected, actual)
            }
        })
    }
}

func assertWork(t *testing.T, expected, actual *workstore.Work) {
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
    
    if expected.Start != actual.Start {
        t.Errorf("expected Start %s but got %s", expected.Start, actual.Start)
    }
    
    if expected.Stop != actual.Stop {
        t.Errorf("expected Stop  but got ", expected.Stop, actual.Stop)
    }
    

    if actual.CreatedAt.IsZero() {
        t.Error("created at should be greater than the zero date")
    }

    if actual.ModifiedAt.IsZero() {
        t.Error("modified at should be greater than the zero date")
    }
}

func TestWork_Update(t *testing.T) {
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
        prepare     *workstore.Work
        actual      *workstore.Work
        expected    *workstore.Work
        expectedErr error
    }{
        {
            name:        "work is nil",
            expectedErr: workstore.ErrDataMissing,
        },
        {
            name:        "work has id only",
            actual:      &workstore.Work{
        ID: testingutils.UUIDParse(t, "70957024-cefb-4bc9-910f-32558a0c420e"),
},
                expectedErr: workstore.ErrDataMissing,
        },
        {
            name:        "work has stop only",
            actual:      &workstore.Work{
        Stop: ,
},
                expectedErr: workstore.ErrDataMissing,
        },
        {
            name:        "work has no id",
            actual:      &workstore.Work{
        Start: "PsxnlpKZOEJDFbnKKq8c6PN",
        Stop: ,
},
                expectedErr: workstore.ErrIDMissing,
        },
        {
            name:        "not existing",
            actual:      &workstore.Work{
        ID: testingutils.UUIDParse(t, "83b38677-1abe-4c46-a41c-69852baa6226"),
        Start: "hZ8UnIhelhPZZa0qt4U3GLrKI31lwTNNS8hO81yH2mGz",
},
                expectedErr: sql.ErrNoRows,
        },
        {
            name:        "work has all fields set",
            actual:      &workstore.Work{
        Start: "qBpXNhH8OVO3CarPzeKXiy",
        Stop: ,
},
                prepare: &workstore.Work{
        Start: "FFwlWTAEAW4j5s9aAGGhOgCQgtDVN4BMrevuN9udtTmGAj6tNB",
        Stop: ,
},
                expected: &workstore.Work{
        Start: "qBpXNhH8OVO3CarPzeKXiy",
        Stop: ,
},
        },
        {
            name:        "work has only mandatory fields set",
            actual:      &workstore.Work{
        Start: "pGEnkBOngJTnw6RWXqd4a9n1ylgrxkh2PRVFI",
},
                prepare: &workstore.Work{
        Start: "2LFwf8RFAYUP9r9lgvzcQUnWcqwNE",
},
                expected: &workstore.Work{
        Start: "pGEnkBOngJTnw6RWXqd4a9n1ylgrxkh2PRVFI",
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
                assertWork(t, testCase.expected, testCase.actual)
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

func TestWork_Delete(t *testing.T) {
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
        prepare     *workstore.Work
        expectedErr error
    }{
        {
            name:        "work is nil",
            expectedErr: workstore.ErrIDMissing,
        },
        {
            name:        "work has no ID",
            expectedErr: workstore.ErrIDMissing,
        },
        {
            name: "success",
            prepare: &workstore.Work{
        Start: "inrqrwF",
        Stop: ,
},
        },
        {
            name: "not existing",
            prepare: &workstore.Work{
        ID: testingutils.UUIDParse(t, "464e77b2-ffb3-41dd-b41c-9fdd6166d095"),
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

            actual := &workstore.Work{ID: id}
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

func TestWork_IsValid(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        name     string
        actual   *workstore.Work
        expected bool
    }{
        {
            name:     "work is nil",
            expected: false,
        },
        {
            name:     "work has id only",
            actual:   &workstore.Work{
        ID: testingutils.UUIDParse(t, "25404193-6591-4fcf-8fcf-9c679b7419e3"),
},
            expected: false,
        },
        {
            name:     "work has start only",
            actual:   &workstore.Work{
        Start: "B5O0t42Wb5ZFZVYeiGVBLRXekhj2uuLIFcxo",
},
            expected: true,
        },
        {
            name:     "work has stop only",
            actual:   &workstore.Work{
        Stop: ,
},
            expected: false,
        },
        {
            name:     "mandatory fields only",
            actual:   &workstore.Work{
        Start: "btrkYCknBzRWT4J4GgzRkFSTx3Mv",
},
            expected: true,
        },
        {
            name:     "mandatory fields with id",
            actual:   &workstore.Work{
        ID: testingutils.UUIDParse(t, "25f89be4-a699-4da8-9109-23f02107a275"),
        Start: "JcufqLQan9InpPZeCyV3s23HCrKhe5s8ai5BB3y",
},
            expected: true,
        },
        {
            name:     "all fields",
            actual:   &workstore.Work{
        ID: testingutils.UUIDParse(t, "99f322fa-f82c-410b-bb0f-821d8ce1d57e"),
        Start: "shCVpv3XbiI",
        Stop: ,
},
            expected: true,
        },
        {
            name:     "all fields without id",
            actual:   &workstore.Work{
        Start: "LMssJ6RsRRjdvC3",
        Stop: ,
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
