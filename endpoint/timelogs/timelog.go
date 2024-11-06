package timelogs

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

func (t *timelog) delete(writer http.ResponseWriter, request *http.Request) {
	log := t.svc.NewLogForRequestID(request.Context())
	response := smis.Response{Log: log}

	vars := mux.Vars(request)
	id, ok := vars["id"]
	if !ok {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "",
			External:   "no id defined",
			Internal:   "no id defined",
			Details:    nil,
		})

		return
	}

	idParsed, err := uuid.Parse(id)
	if err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "",
			External:   "no id defined",
			Internal:   "id not a uuid",
			Details:    nil,
		})

		return
	}

	mapper := timelogmapper.New(t.db)
	if err := mapper.Delete(request.Context(), idParsed); err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusInternalServerError,
			Code:       "",
			External:   "failed to delete timelog",
			Internal:   err.Error(),
			Details:    nil,
		})

		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
