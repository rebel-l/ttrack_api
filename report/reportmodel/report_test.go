package reportmodel_test

import (
	"testing"
	"time"

	"github.com/rebel-l/ttrack_api/report/reportmodel"
)

func TestNewReport(t *testing.T) {
	t.Parallel()

	report := reportmodel.NewReport(2024)

	if report.Year != 2024 {
		t.Errorf("Year expected 2024, got %d", report.Year)
	}

	expectedFirstDayOfYear := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !report.FirstDayOfYear.Equal(expectedFirstDayOfYear) {
		t.Errorf("FirstDayOfYear expected %q, got %q", expectedFirstDayOfYear, report.FirstDayOfYear)
	}

	expectedLastDayOfYear := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	if !report.LastDayOfYear.Equal(expectedLastDayOfYear) {
		t.Errorf("LastDayOfYear expected %q, got %q", expectedLastDayOfYear, report.LastDayOfYear)
	}
}

func TestReport_Calculate(t *testing.T) {
	t.Parallel()

	report := reportmodel.NewReport(2024)
	if err := report.Calculate(); err != nil {
		t.Fatalf("Calculate error: %s", err)
	}

	if report.DaysInYear != 366 {
		t.Errorf("DaysInYear expected 366, got %d", report.DaysInYear)
	}

	if report.WorkDaysInYear != 262 {
		t.Errorf("WorkDaysInYear expected 262, got %d", report.WorkDaysInYear)
	}
}
