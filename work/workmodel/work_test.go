package workmodel_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rebel-l/go-utils/testingutils"
	"github.com/rebel-l/ttrack_api/work/workmodel"
)

func TestWork_DecodeJSON(t *testing.T) {
	t.Parallel()

	createdAt, _ := time.Parse(time.RFC3339Nano, "2019-12-31T03:36:57.9167778+01:00")
	modifiedAt, _ := time.Parse(time.RFC3339Nano, "2020-01-01T15:44:57.9168378+01:00")
	start, _ := time.Parse(time.RFC3339Nano, "2022-01-03T09:00:00.9168378+01:00")
	stop, _ := time.Parse(time.RFC3339Nano, "2022-01-03T13:00:00.9168378+01:00")

	testCases := []struct {
		name        string
		actual      *workmodel.Work
		json        io.Reader
		expected    *workmodel.Work
		expectedErr error
	}{
		{
			name: "model is nil",
		},
		{
			name:        "no JSON format",
			actual:      &workmodel.Work{},
			json:        bytes.NewReader([]byte("no JSON")),
			expected:    &workmodel.Work{},
			expectedErr: workmodel.ErrDecodeJSON,
		},
		{
			name:   "success",
			actual: &workmodel.Work{},
			json: bytes.NewReader([]byte(`
                {
    "ID": "8bccabfd-07eb-4df4-8ad0-e18a86b16a04",
    "Start": "2022-01-03T09:00:00.9168378+01:00",
    "Stop": "2022-01-03T13:00:00.9168378+01:00",
    "CreatedAt": "2019-12-31T03:36:57.9167778+01:00",
    "ModifiedAt": "2020-01-01T15:44:57.9168378+01:00"
}
            `)),
			expected: &workmodel.Work{
				ID:         testingutils.UUIDParse(t, "8bccabfd-07eb-4df4-8ad0-e18a86b16a04"),
				Start:      start,
				Stop:       stop,
				CreatedAt:  createdAt,
				ModifiedAt: modifiedAt,
			},
		},
		{
			name:     "empty json",
			actual:   &workmodel.Work{},
			json:     bytes.NewReader([]byte("{}")),
			expected: &workmodel.Work{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.actual.DecodeJSON(testCase.json)
			if !errors.Is(err, testCase.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", testCase.expectedErr, err)

				return
			}

			assertWork(t, testCase.expected, testCase.actual)
		})
	}
}

func assertWork(t *testing.T, expected, actual *workmodel.Work) {
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
		t.Errorf("expected Stop %s but got %s", expected.Stop, actual.Stop)
	}

	if !expected.CreatedAt.Equal(actual.CreatedAt) {
		t.Errorf("expected created at '%s' but got '%s'", expected.CreatedAt.String(), actual.CreatedAt.String())
	}

	if !expected.ModifiedAt.Equal(actual.ModifiedAt) {
		t.Errorf("expected modified at '%s' but got '%s'", expected.ModifiedAt.String(), actual.ModifiedAt.String())
	}
}
