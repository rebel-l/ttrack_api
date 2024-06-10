package publicholidaymapper

import (
	"context"
	"fmt"
	"time"

	"github.com/rebel-l/ttrack_api/timelog/timelogmapper"

	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaystore"
)

func (m *Mapper) LoadAll(ctx context.Context) (publicholidaymodel.PublicHolidays, error) {
	s := &publicholidaystore.PublicHolidays{}

	if err := s.Load(ctx, m.db, ""); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFromDB, err)
	}

	models := make(publicholidaymodel.PublicHolidays)
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

	for i, year := range tYears {
		y := publicholidaymodel.Year(year)
		if _, ok := models[y]; !ok {
			models[y] = []*publicholidaymodel.PublicHoliday{}
		}

		// add a year in the future
		if i == len(tYears)-1 && year == time.Now().Year() {
			models[publicholidaymodel.Year(year+1)] = []*publicholidaymodel.PublicHoliday{}
		}
	}

	return models, nil
}
