package timelogs

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rebel-l/smis"
	"github.com/rebel-l/ttrack_api/timelog/timelogmapper"
	"github.com/sirupsen/logrus"
)

func (t *timelog) loadByRange(writer http.ResponseWriter, request *http.Request) {
	log := t.svc.NewLogForRequestID(request.Context())
	response := smis.Response{Log: log}

	if writer == nil || request == nil {
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

	vars := mux.Vars(request)
	start, ok := vars["start"]
	if !ok {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "",
			External:   "no start defined",
			Internal:   "no start defined",
			Details:    nil,
		})

		return
	}

	stop, ok := vars["stop"]
	if !ok {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "",
			External:   "no stop defined",
			Internal:   "no stop defined",
			Details:    nil,
		})

		return
	}

	mapper := timelogmapper.New(t.db)

	model, err := mapper.LoadByDateRange(request.Context(), start, stop)
	if err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusInternalServerError,
			Code:       "",
			External:   "failed to load time logs",
			Internal:   "failed to load time logs",
			Details:    err,
		})

		return
	}

	response.WriteJSON(writer, http.StatusOK, model)
}
