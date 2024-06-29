package reportmodel

import (
	"fmt"
	"strings"
	"time"

	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
	"github.com/rebel-l/ttrack_api/timelog/timelogmodel"
	"golang.org/x/exp/maps"
)

const oneDay = 24 * time.Hour

// Report represents all the values to present a proper yearly report of timelogs.
type Report struct {
	Year                     int                 `json:"Year"`
	Days                     int                 `json:"Days"`
	WorkDays                 int                 `json:"WorkDays"`
	DaysOnWeekend            int                 `json:"DaysOnWeekend"`
	PublicHolidays           int                 `json:"PublicHolidays"`
	PublicHolidaysOnWorkdays int                 `json:"PublicHolidaysOnWorkdays"`
	FirstDay                 time.Time           `json:"FirstDay"`
	LastDay                  time.Time           `json:"LastDay"`
	WorkDaysPerReason        map[string]uint32   `json:"WorkDaysPerReason"`
	WorksDaysPerLocation     map[string]uint32   `json:"WorksDaysPerLocation"`
	Warnings                 map[string][]string `json:"Warnings"`
}

type Summary struct {
	Timelogs map[string][]struct {
		Location  string
		DaysCount int
	} `json:"Timelogs"`
	Warnings struct {
		Message string
		Day     string
		Details []timelogmodel.Timelogs
	}
}

// NewReport returns you a Report struct initialized by a given year. Based on the year it calculates first and last
// day of the year.
func NewReport(year int) *Report {
	firstDayOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	return &Report{
		Year:                 year,
		FirstDay:             firstDayOfYear,
		LastDay:              lastDayOfYear,
		Days:                 0,
		WorkDays:             0,
		WorkDaysPerReason:    make(map[string]uint32),
		WorksDaysPerLocation: make(map[string]uint32),
		Warnings:             make(map[string][]string),
	}
}

// Calculate fills all values for the report based on the FirstDay and LastDay.
func (r *Report) Calculate(publicHolidays publicholidaymodel.PublicHolidays, timelogs timelogmodel.Timelogs) error {
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

	workdayReason := make(map[string]map[string]any)   // key 1 = day, key 2 = reason
	workdayLocation := make(map[string]map[string]any) // key 1 = day, key 2 = location
	for _, timelog := range timelogs {
		keyDay := timelog.Start.Format(time.DateOnly)

		if timelog.Stop == nil || timelog.Stop.IsZero() {
			if _, ok := r.Warnings[keyDay]; !ok {
				r.Warnings[keyDay] = make([]string, 0)
			}

			r.Warnings[keyDay] = append(r.Warnings[keyDay], "no stop time")

			continue
		}

		if timelog.Reason == timelogmodel.ReasonWork {
			if _, ok := workdayLocation[keyDay]; !ok {
				workdayLocation[keyDay] = make(map[string]any)
				workdayLocation[keyDay][timelog.Location] = true
			}
		}

		if timelog.Reason != timelogmodel.ReasonBreak {
			if _, ok := workdayReason[keyDay]; !ok {
				workdayReason[keyDay] = make(map[string]any)
				workdayReason[keyDay][timelog.Reason] = true
			} else {
				workdayReason[keyDay][timelog.Reason] = true
			}
		}
	}

	for keyDay, reasons := range workdayReason {
		if len(reasons) > 1 {
			if _, ok := r.Warnings[keyDay]; !ok {
				r.Warnings[keyDay] = make([]string, 0)
			}

			r.Warnings[keyDay] = append(r.Warnings[keyDay], fmt.Sprintf("too many reasons: %q", strings.Join(maps.Keys(reasons), ", ")))
		}

		for reason, _ := range reasons {
			r.WorkDaysPerReason[reason]++
		}
	}

	for _, locations := range workdayLocation {
		for location, _ := range locations {
			r.WorksDaysPerLocation[location]++
		}
	}

	return nil
}

func workday(date time.Time) bool {
	if date.Weekday() > 0 && date.Weekday() < 6 {
		return true
	}

	return false
}
