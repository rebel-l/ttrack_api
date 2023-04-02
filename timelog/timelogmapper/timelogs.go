package timelogmapper

import (
	"context"
	"fmt"

	"github.com/rebel-l/ttrack_api/timelog/timelogmodel"
	"github.com/rebel-l/ttrack_api/timelog/timelogstore"
)

func (m *Mapper) LoadByDateRange(ctx context.Context, start, stop string) (timelogmodel.Timelogs, error) {
	s := &timelogstore.Timelogs{}

	w := "start >= ? AND (stop < ? OR stop IS NULL)"

	if err := s.Load(ctx, m.db, w, start, stop); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFromDB, err)
	}

	tls := timelogmodel.Timelogs{}
	for _, v := range *s {
		tls = append(tls, StoreToModel(v))
	}

	return tls, nil
}
