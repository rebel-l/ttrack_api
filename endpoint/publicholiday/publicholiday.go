package publicholiday

import (
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymapper"
	"io"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
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
	//
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

	// TODO: implement
}
