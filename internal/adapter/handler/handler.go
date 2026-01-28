// Package handler TODO
package handler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/domain"
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

	h.GET("/temperature/:postalCode", h.GetTemperature)

	return h
}

// GetTemperature TODO
func (h *Handler) GetTemperature(ctx *gin.Context) {
	postalCode, err := h.getPostalCode(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	location, err := h.ag.GetAddress(ctx, postalCode)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	c, err := h.tg.GetTemperature(ctx, location)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	f := c*1.8 + 32
	k := c + 273

	ctx.JSON(http.StatusOK, Response{TempC: c, TempF: f, TempK: k})
}

func (h *Handler) getPostalCode(ctx *gin.Context) (string, error) {
	postalCode := ctx.Param("postalCode")
	err := validation.Validate(postalCode,
		validation.Required,
		validation.Length(8, 8),
		validation.Match(regexp.MustCompile(`^\d+$`)),
	)
	if err != nil {
		return "", fmt.Errorf("validating postal code: %w", err)
	}
	return postalCode, nil
}

func (h *Handler) errorMiddleware(ctx *gin.Context) {

	ctx.Next()

	if ctx.Errors == nil {
		return
	}

	err := ctx.Errors.Last()

	var statusCode int
	switch {
	case errors.As(err, &validation.ErrorObject{}):
		statusCode = http.StatusBadRequest
	default:
		statusCode = http.StatusInternalServerError
	}

	ctx.JSON(statusCode, Err{Error: err.Error()})
}

// Response TODO
type Response struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

// Err TODO
type Err struct {
	Error string `json:"error"`
}
