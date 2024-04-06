package timelogmapper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rebel-l/ttrack_api/timelog/timelogmodel"
	"github.com/rebel-l/ttrack_api/timelog/timelogstore"
)

// GetUniqueYears returns a list of years extracted from the time logs. These years are unique.
func (m *Mapper) GetUniqueYears(ctx context.Context) (timelogmodel.UniqueYears, error) {
	var s timelogstore.UniqueYears

	if err := s.Get(ctx, m.db); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFromDB, err)
	}

	var res timelogmodel.UniqueYears
	for _, v := range s {
		converted, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("%w: %q could not be converted to integer", ErrConvert, v)
		}

		res = append(res, converted)
	}

	return res, nil
}
