package handler

import (
	"fmt"
	"strconv"
	"time"

	"echo_rest_api/pkg/internal"
	"echo_rest_api/pkg/model"

	"github.com/labstack/echo/v4"
)

const (
	gate = "gate"
	none = "none"
	hour = "hour"
)

// GET

// Get stats data for a given location
func (h *Handler) StatsGetData(c echo.Context) (err error) {
	// Get input query parameters
	qp := c.QueryParams()

	// Decode the ID into a MongoDB format
	id, err := internal.DecodeQueryParameterID(qp, "id")
	if err != nil {
		return
	}

	// Parse interval dates
	s, err := time.Parse(time.RFC3339, qp.Get("start"))
	if err != nil {
		return internal.NewBackendError(internal.ErrBEQPInvalidDateTime, nil, 1)
	}
	e, err := time.Parse(time.RFC3339, qp.Get("end"))
	if err != nil {
		return internal.NewBackendError(internal.ErrBEQPInvalidDateTime, nil, 1)
	}

	// Parse timezone
	tRaw := qp.Get("timezone")
	t, err := strconv.Atoi(tRaw)
	if err != nil || t < -12 || t > 12 {
		if err == nil {
			err = fmt.Errorf("invalid timezone offset value: %d", t)
		}
		return internal.NewBackendError(internal.ErrBEQPInvalidTimezone, err, 1)
	}

	// Parse location
	l := qp.Get("location")
	if l == "" || !internal.InSlice(l, []string{gate, "space"}) {
		return internal.NewBackendError(internal.ErrBEQPInvalidLocation, nil, 1)
	}

	// Parse chart type
	ct := qp.Get("chartType")
	if ct == "" || !internal.InSlice(ct, []string{"area", "line", "spline", "column", "stackedColumn", "bar"}) {
		return internal.NewBackendError(internal.ErrBEQPInvalidChartType, nil, 1)
	}

	// Parse interval type
	it := qp.Get("intervalType")
	if it == "" || !internal.InSlice(it, []string{none, hour, "day", "month", "year"}) {
		return internal.NewBackendError(internal.ErrBEQPInvalidIntervalType, nil, 1)
	}

	// Parse isInside
	var in bool
	inRaw := qp.Get("isInside")
	if l == gate {
		if inRaw == "" {
			return internal.NewBackendError(internal.ErrBEQPInvalidIsInside, nil, 1)
		}

		in, err = strconv.ParseBool(inRaw)
		if err != nil {
			return internal.NewBackendError(internal.ErrBEQPInvalidIsInside, nil, 1)
		}
	}

	// Check if trying to get raw data from a gate
	if it == none && l == gate {
		return internal.NewBackendError(internal.ErrBEQPNoRawOnGate, nil, 1)
	}

	// Build format for X axis
	var vf string
	switch it {
	case none:
		it = hour
		fallthrough
	case "hour":
		vf = "DD-MM-YYYY HH:mm:ss"
	case "day":
		vf = "DD-MM-YYYY"
	case "month":
		vf = "MM-YYY"
	case "year":
		vf = "YYYY"
	}

	// Correct start/end dates for timezone offset
	s = s.Add(time.Hour * time.Duration(-t))
	e = e.Add(time.Hour * time.Duration(-t))

	// Init the stats model for the data
	sq := &model.Stats{
		ID:        id,
		Location:  l,
		ChartType: ct,
		IsInside:  in,
		Start:     s,
		End:       e,
		Timezone:  t,
		AxisX: model.StatsAxisX{
			ValueFormat:  vf,
			IntervalType: it,
		},
		Data: []model.StatsData{},
	}

	// Retrieve the statistics
	if sq.Location == gate {
		if err = model.StatsGetGate(h.DB, sq); err != nil {
			return
		}
	} else if sq.Location == "space" {
		if err = model.StatsGetSpace(h.DB, sq); err != nil {
			return
		}
	}

	return HTTPSuccess(c, sq)
}
