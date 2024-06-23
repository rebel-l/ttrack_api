package publicholiday

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymapper"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymodel"
	"github.com/sirupsen/logrus"
)

type publicHoliday struct {
	db  *sqlx.DB
	svc *smis.Service
}

func (p *publicHoliday) loadAll(writer http.ResponseWriter, request *http.Request) {
	log := p.svc.NewLogForRequestID(request.Context())
	response := smis.Response{Log: log}

	if writer == nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "",
			External:   "request could not be handled",
			Internal:   "writer is nil",
			Details:    nil,
		})

		return
	}

	defer func(log logrus.FieldLogger, c ...io.Closer) {
		for _, v := range c {
			if err := v.Close(); err != nil {
				log.Warnf("failed to close: %v", err)
			}
		}
	}(log, request.Body)

	mapper := publicholidaymapper.New(p.db)

	model, err := mapper.LoadAll(request.Context())
	if err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusInternalServerError,
			Code:       "PHL-ALL",
			External:   "failed to load public holidays",
			Internal:   "failed to load public holidays",
			Details:    err,
		})

		return
	}

	response.WriteJSON(writer, http.StatusOK, model)
}

func (p *publicHoliday) save(writer http.ResponseWriter, request *http.Request) {
	log := p.svc.NewLogForRequestID(request.Context())
	response := smis.Response{Log: log}

	if writer == nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "PHL-SAVE",
			External:   "request could not be handled",
			Internal:   "writer is nil",
			Details:    nil,
		})

		return
	}

	defer func(log logrus.FieldLogger, c ...io.Closer) {
		for _, v := range c {
			if err := v.Close(); err != nil {
				log.Warnf("failed to close: %v", err)
			}
		}
	}(log, request.Body)

	var models []*publicholidaymodel.PublicHoliday // nolint: exhaustivestruct
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&models); err != nil {
		response.WriteJSONError(writer, smis.ErrResponseJSONConversion.WithDetails(err)) // TODO: maybe custom error?
		return
	}

	for i, v := range models {
		var err error

		mapper := publicholidaymapper.New(p.db)
		models[i], err = mapper.Save(request.Context(), v)
		if err != nil {
			response.WriteJSONError(writer, smis.ErrResponseJSONConversion.WithDetails(err)) // TODO: maybe custom error?
			return
		}
	}

	model := make(publicholidaymodel.PublicHolidaysByYear)
	if len(models) > 0 {
		year := publicholidaymodel.Year(models[0].Day.Year())
		model[year] = models
	}

	response.WriteJSON(writer, http.StatusOK, model)
}
