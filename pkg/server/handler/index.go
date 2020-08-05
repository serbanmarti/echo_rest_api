package handler

import (
	"net/http"
	"time"

	"echo_rest_api/pkg/internal"

	"github.com/labstack/echo/v4"
)

type (
	Tester struct {
		Found bool
		Key   string
		Value interface{}
		Count int
	}
)

func (h *Handler) Index(c echo.Context) error {
	return c.JSON(http.StatusOK, "Service alive!")
}

func (h *Handler) CacheTest(c echo.Context) error {
	// Get input query parameters
	qp := c.QueryParams()

	// Parse query parameters
	k := qp.Get("key")
	v := qp.Get("value")
	if k == "" {
		return internal.NewBackendError(internal.ErrBEQPMissing, nil, 1)
	}

	// Instantiate a tester object
	var t *Tester

	// Check the cache for the key
	value, exists := h.Cache.Get(k)

	if !exists {
		if v == "" {
			return internal.NewBackendError(internal.ErrBEQPMissing, nil, 1)
		}

		// Save the key-value pair
		h.Cache.SetWithTTL(k, v, 10*time.Second)

		// Form the return object
		t = &Tester{
			Found: false,
			Key:   k,
			Value: v,
			Count: h.Cache.Count(),
		}
	} else {
		// Form the return object
		t = &Tester{
			Found: true,
			Key:   k,
			Value: value,
			Count: h.Cache.Count(),
		}
	}

	return HTTPSuccess(c, t)
}
