package reportmodel

import (
	"time"

	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
)

const oneDay = 24 * time.Hour

// Report represents all the values to present a proper yearly report of timelogs.
type Report struct {
	Year                     int       `json:"Year"`
	Days                     int       `json:"Days"`
	WorkDays                 int       `json:"WorkDays"`
	DaysOnWeekend            int       `json:"DaysOnWeekend"`
	PublicHolidays           int       `json:"PublicHolidays"`
	PublicHolidaysOnWorkdays int       `json:"PublicHolidaysOnWorkdays"`
	FirstDay                 time.Time `json:"FirstDay"`
	LastDay                  time.Time `json:"LastDay"`
}

// NewReport returns you a Report struct initialized by a given year. Based on the year it calculates first and last
// day of the year.
func NewReport(year int) *Report {
	firstDayOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	return &Report{
		Year:     year,
		FirstDay: firstDayOfYear,
		LastDay:  lastDayOfYear,
		Days:     0,
		WorkDays: 0,
	}
}

// Calculate fills all values for the report based on the FirstDay and LastDay.
func (r *Report) Calculate(publicHolidays publicholidaymodel.PublicHolidays) error {
	r.PublicHolidays = len(publicHolidays)
	for _, v := range publicHolidays {
		if workday(v.Day) {
			r.PublicHolidaysOnWorkdays++
		}
	}

	day := r.FirstDay
	for day.Before(r.LastDay) {
		r.Days++

		if workday(day) {
			r.WorkDays++
		} else {
			r.DaysOnWeekend++
		}

		day = day.Add(oneDay)
	}
	r.WorkDays -= r.PublicHolidaysOnWorkdays

	return nil
}

func workday(date time.Time) bool {
	if date.Weekday() > 0 && date.Weekday() < 6 {
		return true
	}

	return false
}
