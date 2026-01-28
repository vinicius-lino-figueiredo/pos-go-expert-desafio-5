// Package domain TODO
package domain

import (
	"context"
	"errors"
)

var (
	// ErrPostalCodeNotFound TODO
	ErrPostalCodeNotFound = errors.New("postal code not found")
)

// AddressGetter TODO
type AddressGetter interface {
	GetAddress(ctx context.Context, postalCode string) (string, error)
}

// TemperatureGetter TODO
type TemperatureGetter interface {
	GetTemperature(ctx context.Context, location string) (float64, error)
}
