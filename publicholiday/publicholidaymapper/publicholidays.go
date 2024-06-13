package publicholidaymapper

import (
	"context"
	"fmt"
	"time"

	"github.com/rebel-l/ttrack_api/timelog/timelogmapper"

	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaystore"
)

func (m *Mapper) LoadAll(ctx context.Context) (publicholidaymodel.PublicHolidaysByYear, error) {
	s := &publicholidaystore.PublicHolidays{}

	if err := s.Load(ctx, m.db, ""); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFromDB, err)
	}

	models := make(publicholidaymodel.PublicHolidaysByYear)
	for _, v := range *s {
		m := StoreToModel(v)

		year := publicholidaymodel.Year(m.Day.Year())
		if _, ok := models[year]; ok {
			models[year] = append(models[year], m)
		} else {
			models[year] = []*publicholidaymodel.PublicHoliday{m}
		}
	}

	tMapper := timelogmapper.New(m.db)
	tYears, err := tMapper.GetUniqueYears(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", timelogmapper.ErrLoadFromDB, err)
	}

	for _, year := range tYears {
		y := publicholidaymodel.Year(year)
		if _, ok := models[y]; !ok {
			models[y] = []*publicholidaymodel.PublicHoliday{}
		}
	}

	// add a year in the future
	nextYear := publicholidaymodel.Year(time.Now().Year() + 1)
	if _, ok := models[nextYear]; !ok {
		models[nextYear] = []*publicholidaymodel.PublicHoliday{}
	}

	return models, nil
}
