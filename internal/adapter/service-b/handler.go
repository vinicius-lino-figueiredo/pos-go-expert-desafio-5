// Package serviceb TODO
package serviceb

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Handler TODO
type Handler struct {
	*gin.Engine
	ag domain.AddressGetter
	tg domain.TemperatureGetter
}

// NewHandler TODO
func NewHandler(ag domain.AddressGetter, tg domain.TemperatureGetter) http.Handler {
	h := &Handler{Engine: gin.New(), ag: ag, tg: tg}

	h.Use(h.errorMiddleware)

	h.POST("/temperature", h.GetTemperature)

	return h
}

// GetTemperature TODO
func (h *Handler) GetTemperature(ctx *gin.Context) {
	propagator := otel.GetTextMapPropagator()
	reqCtx := propagator.Extract(ctx.Request.Context(), propagation.HeaderCarrier(ctx.Request.Header))

	reqCtx, span := otel.Tracer("service-b").Start(reqCtx, "handle-temperature")
	defer span.End()

	postalCode, err := h.getPostalCode(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	location, err := h.ag.GetAddress(reqCtx, postalCode)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	c, err := h.tg.GetTemperature(reqCtx, location)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	f := c*1.8 + 32
	k := c + 273

	ctx.JSON(http.StatusOK, Response{City: location, TempC: c, TempF: f, TempK: k})
}

func (h *Handler) getPostalCode(ctx *gin.Context) (string, error) {
	var body map[string]any
	if err := ctx.BindJSON(&body); err != nil {
		return "", domain.ErrInvalidZipCode
	}

	cep, _ := body["cep"].(string)
	if len(cep) != 8 {
		return "", domain.ErrInvalidZipCode
	}

	return cep, nil
}

func (h *Handler) errorMiddleware(ctx *gin.Context) {

	ctx.Next()

	if ctx.Errors == nil {
		return
	}

	err := ctx.Errors.Last()

	var statusCode int
	switch {
	case errors.Is(err, domain.ErrPostalCodeNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidZipCode):
		statusCode = http.StatusUnprocessableEntity
	default:
		statusCode = http.StatusInternalServerError
	}

	ctx.JSON(statusCode, Err{Error: err.Error()})
}

// Response TODO
type Response struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

// Err TODO
type Err struct {
	Error string `json:"error"`
}
