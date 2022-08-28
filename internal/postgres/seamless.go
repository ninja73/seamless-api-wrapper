package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"seamless-api-wrapper/internal/model"
	"time"
)

var (
	ErrIllegalCurrencyCode = errors.New("ErrIllegalCurrencyCode")
	ErrNotEnoughMoneyCode  = errors.New("ErrNotEnoughMoneyCode")
)

type SeamlessService struct {
	db *sqlx.DB
}

func NewSeamlessService(db *sqlx.DB) *SeamlessService {
	return &SeamlessService{db: db}
}

func (s *SeamlessService) Balance(ctx context.Context, playerName, currency string) (*model.Balance, error) {
	var balance model.Balance
	err := s.db.GetContext(ctx, &balance, "SELECT * FROM balances WHERE player_name = $1 AND currency = $2 LIMIT 1",
		playerName, currency)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (s *SeamlessService) Transaction(ctx context.Context, playerName, currency string, transaction *model.Transaction) (*model.Balance, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var balance model.Balance
	err = tx.Get(&balance, `SELECT *  FROM balances  WHERE player_name = $1 AND currency = $2 LIMIT 1 FOR UPDATE`,
		playerName, currency)

	if errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return nil, ErrIllegalCurrencyCode
	}

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	balance.Amount -= transaction.Withdraw
	if balance.Amount < 0 {
		tx.Rollback()
		return nil, ErrNotEnoughMoneyCode
	}

	balance.Amount += transaction.Deposit

	if transaction.ChargeFreeRounds != nil {
		freeRoundLeft := 0
		if balance.FreeRoundLeft != nil {
			freeRoundLeft = *balance.FreeRoundLeft
		}

		freeRoundLeft -= *transaction.ChargeFreeRounds

		balance.FreeRoundLeft = &freeRoundLeft
	}

	_, err = tx.NamedExec("UPDATE balance SET amount = :amount, free_round_left = :free_round_left WHERE id = :id", &balance)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	transaction.BalanceID = balance.ID

	now := time.Now()

	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	insertTransactionQuery := `INSERT INTO transactions(
     balance_id,  withdraw, deposit, transaction_ref,
     source, reason, session_id, session_alternative_id,
     bonus_id, charge_free_rounds, created_at, updated_at
    ) VALUES (
	 :balance_id, :withdraw, :deposit, :transaction_ref,
	 :source, :reason, :session_id, :session_alternative_id,
	 :bonus_id, :charge_free_rounds, :created_at, :updated_at
	) RETURNING id`

	if _, err = tx.NamedExec(insertTransactionQuery, transaction); err != nil {
		tx.Rollback()
		return nil, err
	}

	return &balance, tx.Commit()
}

func (s *SeamlessService) Rollback(ctx context.Context, transaction *model.Transaction) error {
	return nil
}
