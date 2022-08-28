package model

import "time"

type Balance struct {
	ID                       int
	PlayerName               string `db:"player_name"`
	Currency                 string
	Amount                   int
	GameID                   *string   `db:"game_id"`
	LastSessionID            *string   `db:"last_session_id"`
	LastSessionAlternativeID *string   `db:"last_session_alternative_id"`
	FreeRoundLeft            *int      `db:"free_round_left"`
	CreatedAt                time.Time `db:"created_at"`
	UpdatedAt                time.Time `db:"updated_at"`
}
