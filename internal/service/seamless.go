package service

import (
	"context"
	"seamless-api-wrapper/internal/model"
)

type SeamlessService interface {
	Balance(ctx context.Context, playerName, currency string) (*model.Balance, error)
	Transaction(ctx context.Context, playerName, currency string, transaction *model.Transaction) (*model.Balance, error)
	Rollback(ctx context.Context, playerName string, transaction *model.Transaction) error
}
