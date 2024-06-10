package publicholidaymodel_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rebel-l/go-utils/testingutils"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
)

func TestPublicHoliday_DecodeJSON(t *testing.T) {
	t.Parallel()

	createdAt, _ := time.Parse(time.RFC3339Nano, "2019-12-31T03:36:57.9167778+01:00")
	modifiedAt, _ := time.Parse(time.RFC3339Nano, "2020-01-01T15:44:57.9168378+01:00")

	now := time.Date(2024, 5, 10, 17, 0, 0, 0, time.UTC)

	testCases := []struct {
		name        string
		actual      *publicholidaymodel.PublicHoliday
		json        io.Reader
		expected    *publicholidaymodel.PublicHoliday
		expectedErr error
	}{
		{
			name: "model is nil",
		},
		{
			name:        "no JSON format",
			actual:      &publicholidaymodel.PublicHoliday{},
			json:        bytes.NewReader([]byte("no JSON")),
			expected:    &publicholidaymodel.PublicHoliday{},
			expectedErr: publicholidaymodel.ErrDecodeJSON,
		},
		{
			name:   "success",
			actual: &publicholidaymodel.PublicHoliday{},
			json: bytes.NewReader([]byte(`
                {
    "ID": "db2cfef7-d8c0-458d-a72c-de5bdc1a5b62",
    "Day": "2024-05-10T17:00:00Z",
    "Name": "YlNorwOEK73cSeRpoUVkyH7LlR1YioOY",
    "HalfDay": true,
    "CreatedAt": "2019-12-31T03:36:57.9167778+01:00",
    "ModifiedAt": "2020-01-01T15:44:57.9168378+01:00"
}
            `)),
			expected: &publicholidaymodel.PublicHoliday{
				ID:         testingutils.UUIDParse(t, "db2cfef7-d8c0-458d-a72c-de5bdc1a5b62"),
				Day:        now,
				Name:       "YlNorwOEK73cSeRpoUVkyH7LlR1YioOY",
				HalfDay:    true,
				CreatedAt:  createdAt,
				ModifiedAt: modifiedAt,
			},
		},
		{
			name:     "empty json",
			actual:   &publicholidaymodel.PublicHoliday{},
			json:     bytes.NewReader([]byte("{}")),
			expected: &publicholidaymodel.PublicHoliday{},
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

			assertPublicHoliday(t, testCase.expected, testCase.actual)
		})
	}
}

func assertPublicHoliday(t *testing.T, expected, actual *publicholidaymodel.PublicHoliday) {
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

	if !expected.CreatedAt.Equal(actual.CreatedAt) {
		t.Errorf("expected created at '%s' but got '%s'", expected.CreatedAt.String(), actual.CreatedAt.String())
	}

	if !expected.ModifiedAt.Equal(actual.ModifiedAt) {
		t.Errorf("expected modified at '%s' but got '%s'", expected.ModifiedAt.String(), actual.ModifiedAt.String())
	}
}
