package timelogs

import (
	"io"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
	"github.com/rebel-l/ttrack_api/timelog/timelogmapper"
	"github.com/rebel-l/ttrack_api/timelog/timelogmodel"
	"github.com/sirupsen/logrus"
)

type timelog struct {
	db  *sqlx.DB
	svc *smis.Service
}

// Init initializes the endpoints to log times.
func Init(svc *smis.Service, db *sqlx.DB) error {
	endpoint := &timelog{db: db, svc: svc}

	_, err := svc.RegisterEndpoint("/timgelogs", http.MethodPut, endpoint.upsert)

	return err // nolint: wrapcheck
}

func (t *timelog) upsert(writer http.ResponseWriter, request *http.Request) {
	log := t.svc.NewLogForRequestID(request.Context())
	response := smis.Response{Log: log}

	if writer == nil || request == nil || request.Body == nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "",
			External:   "request had no data",
			Internal:   "writer or request nil",
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

	model := &timelogmodel.Timelog{} // nolint: exhaustivestruct
	if err := model.DecodeJSON(request.Body); err != nil {
		response.WriteJSONError(writer, smis.ErrResponseJSONConversion.WithDetails(err))

		return
	}

	if err := model.Validate(); err != nil {
		response.WriteJSONError(writer, smis.Error{ // nolint: exhaustivestruct
			StatusCode: http.StatusBadRequest,
			Code:       "VALIDATION",
			External:   err.Error(),
		})

		return
	}

	mapper := timelogmapper.New(t.db)

	model, err := mapper.Save(request.Context(), model)
	if err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusInternalServerError,
			Code:       "SAVE",
			External:   "failed to save timelog",
			Internal:   "failed to save timelog",
			Details:    err,
		})

		return
	}

	response.WriteJSON(writer, http.StatusOK, model)
}
