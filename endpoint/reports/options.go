package reports

import (
	"io"
	"net/http"

	"github.com/rebel-l/smis"
	"github.com/rebel-l/ttrack_api/timelog/timelogmapper"
	"github.com/sirupsen/logrus"
)

func (r *reports) options(writer http.ResponseWriter, request *http.Request) {
	log := r.svc.NewLogForRequestID(request.Context())
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

	mapper := timelogmapper.New(r.db)

	model, err := mapper.GetUniqueYears(request.Context())
	if err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusInternalServerError,
			Code:       "RPT-OPT",
			External:   "failed to load report options",
			Internal:   "failed to load options",
			Details:    err,
		})

		return
	}

	response.WriteJSON(writer, http.StatusOK, model)
}
