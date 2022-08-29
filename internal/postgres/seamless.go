package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"seamless-api-wrapper/internal/model"
	"seamless-api-wrapper/internal/service"
	"time"
)

type SeamlessService struct {
	db *sqlx.DB
}

func NewSeamlessService(db *sqlx.DB) *SeamlessService {
	return &SeamlessService{db: db}
}

func (s *SeamlessService) Balance(ctx context.Context, playerName, currencyCode string) (*model.Balance, error) {
	var balance model.Balance
	err := s.db.GetContext(ctx, &balance, `SELECT 
	    balances.*,
        currencies.id "currency.id",
        currencies.code "currency.code"
	FROM balances 
	LEFT JOIN currencies ON balances.currency_id = currencies.id 
	WHERE balances.player_name = $1 AND currencies.code = $2 LIMIT 1`, playerName, currencyCode)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrNotEnoughMoneyCode
	}

	if err != nil {
		return nil, err
	}

	if balance.Currency.Code == "" {
		return nil, service.ErrIllegalCurrencyCode
	}

	return &balance, nil
}

func (s *SeamlessService) Transaction(ctx context.Context, playerName, currencyCode string, transaction *model.Transaction) (*model.Balance, error) {
	if transaction.Deposit < 0 {
		return nil, service.ErrNegativeDepositCode
	}

	if transaction.Withdraw < 0 {
		return nil, service.ErrNegativeWithdrawalCode
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var balance model.Balance
	err = tx.Get(&balance, `SELECT 
        balances.*,
        currencies.id "currency.id",
        currencies.code "currency.code"
	FROM balances 
	LEFT JOIN currencies ON balances.currency_id = currencies.id 
	WHERE balances.player_name = $1 AND currencies.code = $2 LIMIT 1 
	FOR UPDATE OF balances`, playerName, currencyCode)

	if errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return nil, service.ErrIllegalCurrencyCode
	}

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if balance.Currency.Code == "" {
		return nil, service.ErrIllegalCurrencyCode
	}

	transaction.BalanceID = balance.ID

	now := time.Now()

	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	insertTransactionQuery := `INSERT INTO transactions(
        balance_id,  withdraw, deposit, game_id, transaction_ref,
        source, reason, session_id, session_alternative_id,
        bonus_id, charge_free_rounds, created_at, updated_at
    ) VALUES (
        :balance_id, :withdraw, :deposit, :game_id, :transaction_ref,
        :source, :reason, :session_id, :session_alternative_id,
        :bonus_id, :charge_free_rounds, :created_at, :updated_at
	) RETURNING id`

	query, args, err := tx.BindNamed(insertTransactionQuery, transaction)
	if err != nil {
		return nil, err
	}

	err = tx.Get(&transaction.ID, query, args...)
	if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
		tx.Rollback()
		return &balance, s.checkTransaction(ctx, transaction)
	}

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	balance.Amount -= transaction.Withdraw
	if balance.Amount < 0 {
		tx.Rollback()
		return nil, service.ErrSpendingBudgetExceeded
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

	_, err = tx.NamedExec("UPDATE balances SET amount = :amount, free_round_left = :free_round_left WHERE id = :id", &balance)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &balance, tx.Commit()
}

func (s *SeamlessService) Rollback(ctx context.Context, playerName string, transaction *model.Transaction) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	query, args, err := tx.BindNamed(`INSERT INTO transactions(
        balance_id, withdraw, deposit, game_id, transaction_ref, 
        session_id, session_alternative_id, is_rollback, created_at, updated_at) 
	VALUES (
        :balance_id, :withdraw, :deposit, :game_id, :transaction_ref,
        :session_id, :session_alternative_id, :is_rollback, :created_at, :updated_at
	) 
	ON CONFLICT(transaction_ref) 
	DO UPDATE SET game_id=EXCLUDED.game_id  
	RETURNING id`, transaction)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Get(&transaction.ID, query, args...); err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Get(transaction, "SELECT withdraw, deposit, is_rollback, charge_free_rounds FROM transactions WHERE id = $1", transaction.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if transaction.IsRollback {
		return tx.Commit()
	}

	_, err = tx.Exec(`UPDATE balances 
            SET amount = amount + $1 - $2, 
                free_round_left = coalesce(free_round_left, 0) + coalesce($3, 0) 
            WHERE player_name = $4`,
		transaction.Withdraw, transaction.Deposit, transaction.ChargeFreeRounds, playerName)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE transactions SET is_rollback = $1 WHERE id = $2", true, transaction.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *SeamlessService) checkTransaction(ctx context.Context, transaction *model.Transaction) error {
	err := s.db.GetContext(ctx, transaction, "SELECT * FROM transactions WHERE transaction_ref = $1 LIMIT 1",
		transaction.TransactionRef)
	if err != nil {
		return err
	}

	if transaction.IsRollback {
		return service.ErrTransactionRollback
	}

	return nil
}
