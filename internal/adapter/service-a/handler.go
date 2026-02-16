// Package servicea TODO
package servicea

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/domain"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

// Handler TODO
type Handler struct {
	*gin.Engine
	serviceBURL string
	client      *http.Client
}

// NewHandler TODO
func NewHandler(serviceBURL string) http.Handler {
	h := &Handler{
		Engine:      gin.New(),
		serviceBURL: serviceBURL,
		client:      &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
	}

	h.Use(h.errorMiddleware)

	h.POST("/temperature", h.GetTemperature)

	return h
}

// GetTemperature TODO
func (h *Handler) GetTemperature(ctx *gin.Context) {
	reqCtx, span := otel.Tracer("service-a").Start(ctx.Request.Context(), "forward-to-service-b")
	defer span.End()

	postalCode, err := h.getPostalCode(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	w := bytes.NewBuffer(make([]byte, 0, 64))

	err = json.NewEncoder(w).Encode(map[string]any{"cep": postalCode})
	if err != nil {
		ctx.Error(err)
		return
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, h.serviceBURL, w)
	if err != nil {
		ctx.Error(err)
		return
	}

	res, err := h.client.Do(req)
	if err != nil {
		ctx.Error(err)
		return
	}

	defer res.Body.Close()

	// preferia usar io.Copy, mas é melhor deixar o chi fazer marshal pra o
	// Conten-Type ser setado corretamente
	var response map[string]any
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(res.StatusCode, response)
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
