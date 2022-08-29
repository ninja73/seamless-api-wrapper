package service

import "errors"

var (
	ErrNotEnoughMoneyCode     = errors.New("ErrNotEnoughMoneyCode")
	ErrIllegalCurrencyCode    = errors.New("ErrIllegalCurrencyCode")
	ErrNegativeDepositCode    = errors.New("ErrNegativeDepositCode")
	ErrNegativeWithdrawalCode = errors.New("ErrNegativeWithdrawalCode")
	ErrSpendingBudgetExceeded = errors.New("ErrSpendingBudgetExceeded")
	ErrTransactionRollback    = errors.New("ErrTransactionRollback")
)
