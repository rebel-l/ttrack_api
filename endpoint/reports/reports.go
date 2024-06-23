package reports

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/rebel-l/smis"
	"github.com/rebel-l/ttrack_api/publicholiday/publicholidaymapper"
	"github.com/rebel-l/ttrack_api/report/reportmodel"
	"github.com/sirupsen/logrus"
)

type reports struct {
	db  *sqlx.DB
	svc *smis.Service
}

func (r *reports) reports(writer http.ResponseWriter, request *http.Request) {
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

	vars := mux.Vars(request)
	year, ok := vars["year"]
	if !ok { //nolint: wsl
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "RPT-NOPARAM",
			External:   "no year defined",
			Internal:   "no year defined",
			Details:    nil,
		})

		return
	}

	yearNum, err := strconv.Atoi(year)
	if err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusBadRequest,
			Code:       "RPT-WRONGPARAM",
			External:   "cannot parse year",
			Internal:   "cannot parse year",
			Details:    err,
		})

		return
	}

	mapper := publicholidaymapper.New(r.db)
	publicHolidays, err := mapper.LoadByYear(request.Context(), yearNum)
	if err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusInternalServerError,
			Code:       "RPT-PHL",
			External:   "failed to calculate report",
			Internal:   "failed to load public holidays",
			Details:    err,
		})
	}

	model := reportmodel.NewReport(yearNum)
	if err := model.Calculate(publicHolidays); err != nil {
		response.WriteJSONError(writer, smis.Error{
			StatusCode: http.StatusInternalServerError,
			Code:       "",
			External:   "failed to calculate report",
			Internal:   "failed to calculate report",
			Details:    err,
		})

		return
	}

	response.WriteJSON(writer, http.StatusOK, model)
}
