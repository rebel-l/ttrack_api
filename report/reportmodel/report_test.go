package reportmodel_test

import (
	"testing"
	"time"

	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
	"github.com/rebel-l/ttrack_api/report/reportmodel"
	"github.com/rebel-l/ttrack_api/timelog/timelogmodel"
)

func TestNewReport(t *testing.T) {
	t.Parallel()

	report := reportmodel.NewReport(2024)

	if report.Year != 2024 {
		t.Errorf("Year expected 2024, got %d", report.Year)
	}

	expectedFirstDayOfYear := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !report.FirstDay.Equal(expectedFirstDayOfYear) {
		t.Errorf("FirstDay expected %q, got %q", expectedFirstDayOfYear, report.FirstDay)
	}

	expectedLastDayOfYear := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	if !report.LastDay.Equal(expectedLastDayOfYear) {
		t.Errorf("LastDay expected %q, got %q", expectedLastDayOfYear, report.LastDay)
	}
}

func TestReport_Calculate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		publicHolidays publicholidaymodel.PublicHolidays
		timelogs       timelogmodel.Timelogs
		expected       struct {
			PublicHolidays           int
			PublicHolidaysOnWorkdays int
			Workdays                 int
		}
	}{
		{
			name: "empty public holidays",
			expected: struct {
				PublicHolidays           int
				PublicHolidaysOnWorkdays int
				Workdays                 int
			}{PublicHolidays: 0, PublicHolidaysOnWorkdays: 0, Workdays: 262},
		},
		{
			name: "with public holidays",
			publicHolidays: publicholidaymodel.PublicHolidays{
				{Day: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},  // weekend
				{Day: time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC)}, // workday
			},
			expected: struct {
				PublicHolidays           int
				PublicHolidaysOnWorkdays int
				Workdays                 int
			}{PublicHolidays: 2, PublicHolidaysOnWorkdays: 1, Workdays: 261},
		},
		// TODO: add test cases with timelogs
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			report := reportmodel.NewReport(2024)
			if err := report.Calculate(testCase.publicHolidays, testCase.timelogs); err != nil {
				t.Fatalf("Calculate error: %s", err)
			}

			if report.Year != 2024 {
				t.Errorf("Year expected 2024, got %d", report.Year)
			}

			if report.Days != 366 {
				t.Errorf("Days expected 366, got %d", report.Days)
			}

			if report.WorkDays != testCase.expected.Workdays {
				t.Errorf("WorkDays expected %d, got %d", testCase.expected.Workdays, report.WorkDays)
			}

			if report.DaysOnWeekend != 104 {
				t.Errorf("DaysOnWeekend expected 104, got %d", report.DaysOnWeekend)
			}

			if report.PublicHolidays != testCase.expected.PublicHolidays {
				t.Errorf("PublicHolidays expected %d, got %d", testCase.expected.PublicHolidays, report.PublicHolidays)
			}

			if report.PublicHolidaysOnWorkdays != testCase.expected.PublicHolidaysOnWorkdays {
				t.Errorf("PublicHolidaysOnWorkdays expected %d, got %d", testCase.expected.PublicHolidaysOnWorkdays, report.PublicHolidaysOnWorkdays)
			}

			if !report.FirstDay.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
				t.Errorf("FirstDay expected 01.01.2024, got %q", report.FirstDay)
			}

			if !report.LastDay.Equal(time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)) {
				t.Errorf("LastDay expected 31.12.2024 23:59:59, got %q", report.LastDay)
			}
		})
	}
}
