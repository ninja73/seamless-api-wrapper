package model

import "time"

type Transaction struct {
	ID                   int
	BalanceID            int `db:"balance_id"`
	Withdraw             int
	Deposit              int
	TransactionRef       string  `db:"transaction_ref"`
	GameID               *string `db:"game_id"`
	GameRoundRef         *string `db:"game_round_ref"`
	Source               *string
	Reason               *string
	SessionId            *string   `db:"session_id"`
	SessionAlternativeId *string   `db:"session_alternative_id"`
	BonusId              *string   `db:"bonus_id"`
	ChargeFreeRounds     *int      `db:"charge_free_rounds"`
	IsRollback           bool      `db:"is_rollback"`
	CreatedAt            time.Time `db:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"`
}

type SpinDetails struct {
	ID            int
	TransactionID int    `db:"transaction_id"`
	BetType       string `db:"bet_type"`
	WinType       string `db:"win_type"`
}
