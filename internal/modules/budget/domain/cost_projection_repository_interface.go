package domain

import (
	"context"
	"errors"
)

var ErrCostProjectionNotFound = errors.New("cost projection not found")

type CostProjectionRepositoryInterface interface {
	FindCostProjection(context.Context, string, int, string) (*CostProjection, error)
	SaveCostProjection(context.Context, *CostProjection) (*CostProjection, error)
}
