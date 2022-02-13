package timelogmodel_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rebel-l/go-utils/testingutils"
	"github.com/rebel-l/ttrack_api/timelog/timelogmodel"
)

func TestTimelog_DecodeJSON(t *testing.T) {
	t.Parallel()

	createdAt, _ := time.Parse(time.RFC3339Nano, "2019-12-31T03:36:57.9167778+01:00")
	modifiedAt, _ := time.Parse(time.RFC3339Nano, "2020-01-01T15:44:57.9168378+01:00")

	testTime := testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5218364+01:00")

	testCases := []struct {
		name        string
		actual      *timelogmodel.Timelog
		json        io.Reader
		expected    *timelogmodel.Timelog
		expectedErr error
	}{
		{
			name: "model is nil",
		},
		{
			name:        "no JSON format",
			actual:      &timelogmodel.Timelog{},
			json:        bytes.NewReader([]byte("no JSON")),
			expected:    &timelogmodel.Timelog{},
			expectedErr: timelogmodel.ErrDecodeJSON,
		},
		{
			name:   "success",
			actual: &timelogmodel.Timelog{},
			json: bytes.NewReader([]byte(`
                {
    "ID": "1cbe5ff0-332a-4118-baf3-877cb70e984e",
    "Start": "2022-01-09T22:21:59.5218364+01:00",
    "Stop": "2022-01-09T22:21:59.5218364+01:00",
    "Reason": "jhLQzmBNjLE74Gl",
    "Location": "UGPX68UzOx9oqx",
    "CreatedAt": "2019-12-31T03:36:57.9167778+01:00",
    "ModifiedAt": "2020-01-01T15:44:57.9168378+01:00"
}
            `)),
			expected: &timelogmodel.Timelog{
				ID:         testingutils.UUIDParse(t, "1cbe5ff0-332a-4118-baf3-877cb70e984e"),
				Start:      testingutils.TimeParse("2006-01-02T15:04:05.999999999Z07:00", "2022-01-09T22:21:59.5218364+01:00"),
				Stop:       &testTime,
				Reason:     "jhLQzmBNjLE74Gl",
				Location:   "UGPX68UzOx9oqx",
				CreatedAt:  createdAt,
				ModifiedAt: modifiedAt,
			},
		},
		{
			name:     "empty json",
			actual:   &timelogmodel.Timelog{},
			json:     bytes.NewReader([]byte("{}")),
			expected: &timelogmodel.Timelog{},
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

			assertTimelog(t, testCase.expected, testCase.actual)
		})
	}
}

func assertTimelog(t *testing.T, expected, actual *timelogmodel.Timelog) { // nolint: gocognit
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

	if expected.Stop != nil && actual.Stop != nil && !expected.Stop.Equal(*actual.Stop) ||
		expected.Stop == nil && actual.Stop != nil || expected.Stop != nil && actual.Stop == nil {
		t.Errorf("expected Stop %s but got %s", expected.Stop, actual.Stop)
	}

	if expected.Reason != actual.Reason {
		t.Errorf("expected Reason %s but got %s", expected.Reason, actual.Reason)
	}

	if expected.Location != actual.Location {
		t.Errorf("expected Location %s but got %s", expected.Location, actual.Location)
	}

	if !expected.CreatedAt.Equal(actual.CreatedAt) {
		t.Errorf("expected created at '%s' but got '%s'", expected.CreatedAt.String(), actual.CreatedAt.String())
	}

	if !expected.ModifiedAt.Equal(actual.ModifiedAt) {
		t.Errorf("expected modified at '%s' but got '%s'", expected.ModifiedAt.String(), actual.ModifiedAt.String())
	}
}
