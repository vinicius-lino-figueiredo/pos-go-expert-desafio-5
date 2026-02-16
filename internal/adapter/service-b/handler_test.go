package serviceb_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	serviceb "github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/service-b"
)

type mockAddressGetter struct {
	address string
	err     error
}

func (m *mockAddressGetter) GetAddress(_ context.Context, _ string) (string, error) {
	return m.address, m.err
}

type mockTemperatureGetter struct {
	temp float64
	err  error
}

func (m *mockTemperatureGetter) GetTemperature(_ context.Context, _ string) (float64, error) {
	return m.temp, m.err
}

type HandlerSuite struct {
	suite.Suite
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

func (s *HandlerSuite) TestInvalidPostalCodeTooShort() {
	h := serviceb.NewHandler(&mockAddressGetter{}, &mockTemperatureGetter{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/temperature", strings.NewReader(`{"cep":"123"}`))

	h.ServeHTTP(rec, req)

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

func (s *HandlerSuite) TestInvalidPostalCodeEmpty() {
	h := serviceb.NewHandler(&mockAddressGetter{}, &mockTemperatureGetter{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/temperature", strings.NewReader(`{"cep":""}`))

	h.ServeHTTP(rec, req)

	s.Equal(http.StatusUnprocessableEntity, rec.Code)
}

func (s *HandlerSuite) TestAddressGetterError() {
	ag := &mockAddressGetter{err: errors.New("not found")}
	h := serviceb.NewHandler(ag, &mockTemperatureGetter{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/temperature", strings.NewReader(`{"cep":"01001000"}`))

	h.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *HandlerSuite) TestTemperatureGetterError() {
	ag := &mockAddressGetter{address: "São Paulo"}
	tg := &mockTemperatureGetter{err: errors.New("service unavailable")}
	h := serviceb.NewHandler(ag, tg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/temperature", strings.NewReader(`{"cep":"01001000"}`))

	h.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *HandlerSuite) TestSuccessfulResponse() {
	ag := &mockAddressGetter{address: "São Paulo"}
	tg := &mockTemperatureGetter{temp: 25.0}
	h := serviceb.NewHandler(ag, tg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/temperature", strings.NewReader(`{"cep":"01001000"}`))

	h.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var resp serviceb.Response
	err := json.NewDecoder(rec.Body).Decode(&resp)
	s.NoError(err)
	s.Equal(25.0, resp.TempC)
	s.Equal(25.0*1.8+32, resp.TempF)
	s.Equal(25.0+273, resp.TempK)
}
