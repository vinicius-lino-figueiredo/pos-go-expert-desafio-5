// Package domain TODO
package domain

import (
	"context"
	"errors"
)

var (
	// ErrInvalidZipCode TODO
	ErrInvalidZipCode = errors.New("invalid zipcode")
	// ErrPostalCodeNotFound TODO
	ErrPostalCodeNotFound = errors.New("can not find zipcode")
)

// AddressGetter TODO
type AddressGetter interface {
	GetAddress(ctx context.Context, postalCode string) (string, error)
}

// TemperatureGetter TODO
type TemperatureGetter interface {
	GetTemperature(ctx context.Context, location string) (float64, error)
}
