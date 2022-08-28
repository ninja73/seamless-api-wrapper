package dto

type GetBalanceReq struct {
	CallerId             int     `json:"callerId" validate:"required"`
	PlayerName           string  `json:"playerName" validate:"required"`
	Currency             string  `json:"currency" validate:"required,iso4217"`
	GameId               *string `json:"gameId"`
	SessionId            *string `json:"sessionId"`
	SessionAlternativeId *string `json:"sessionAlternativeId"`
	BonusId              *string `json:"bonusId"`
}

type GetBalanceResp struct {
	Balance        int  `json:"balance" validate:"required"`
	FreeRoundsLeft *int `json:"freeroundsLeft,omitempty"`
}

type SpinDetails struct {
	BetType string `json:"betType"`
	WinType string `json:"winType"`
}

type WithdrawAndDepositReq struct {
	CallerId             int          `json:"callerId" validate:"required"`
	PlayerName           string       `json:"playerName" validate:"required"`
	Withdraw             int          `json:"withdraw" validate:"required"`
	Deposit              int          `json:"deposit" validate:"required"`
	Currency             string       `json:"currency" validate:"required,iso4217"`
	TransactionRef       string       `json:"transactionRef" validate:"required"`
	GameRoundRef         *string      `json:"gameRoundRef"`
	GameId               *string      `json:"gameId"`
	Source               *string      `json:"source"`
	Reason               *string      `json:"reason"`
	SessionId            *string      `json:"sessionId"`
	SessionAlternativeId *string      `json:"sessionAlternativeId"`
	SpinDetails          *SpinDetails `json:"spinDetails"`
	BonusId              *string      `json:"bonusId"`
	ChargeFreeRounds     *int         `json:"chargeFreerounds"`
}

type WithdrawAndDepositResp struct {
	NewBalance     int    `json:"newBalance" validate:"required"`
	TransactionId  string `json:"transactionId" validate:"required"`
	FreeRoundsLeft *int   `json:"freeroundsLeft,omitempty"`
}

type RollbackTransactionReq struct {
	CallerId             int     `json:"callerId" validate:"required"`
	PlayerName           string  `json:"playerName" validate:"required"`
	TransactionRef       string  `json:"transactionRef" validate:"required"`
	GameId               *string `json:"gameId"`
	SessionId            *string `json:"sessionId"`
	SessionAlternativeId *string `json:"sessionAlternativeId"`
	GameRoundRef         *string `json:"roundId"`
}
