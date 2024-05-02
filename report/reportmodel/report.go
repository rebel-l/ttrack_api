package reportmodel

import "time"

const oneDay = 24 * time.Hour

// Report represents all the values to present a proper yearly report of timelogs.
type Report struct {
	Year           int       `json:"Year"`
	DaysInYear     int       `json:"DaysInYear"`
	WorkDaysInYear int       `json:"WorkDaysInYear"`
	FirstDayOfYear time.Time `json:"FirstDayOfYear"`
	LastDayOfYear  time.Time `json:"LastDayOfYear"`
}

// NewReport returns you a Report struct initialized by a given year. Based on the year it calculates first and last
// day of the year.
func NewReport(year int) *Report {
	firstDayOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	return &Report{
		Year:           year,
		FirstDayOfYear: firstDayOfYear,
		LastDayOfYear:  lastDayOfYear,
		DaysInYear:     0,
		WorkDaysInYear: 0,
	}
}

// Calculate fills all values for the report based on the FirstDayOfYear and LastDayOfYear.
func (r *Report) Calculate() error {
	day := r.FirstDayOfYear
	for day.Before(r.LastDayOfYear) {
		r.DaysInYear++

		if workday(day) {
			r.WorkDaysInYear++
		}

		day = day.Add(oneDay)
	}

	return nil
}

func workday(date time.Time) bool {
	if date.Weekday() > 0 && date.Weekday() < 6 {
		return true
	}

	return false
}
