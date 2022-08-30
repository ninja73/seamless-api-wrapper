package seamless

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"seamless-api-wrapper/internal/model"
	"seamless-api-wrapper/internal/rpc"
	"seamless-api-wrapper/internal/service"
	"seamless-api-wrapper/internal/validate"
	"seamless-api-wrapper/package/dto"
	"strconv"
	"time"
)

type Seamless struct {
	seamlessService service.SeamlessService
}

func NewSeamless(seamlessService service.SeamlessService) *Seamless {
	return &Seamless{seamlessService: seamlessService}
}

func (s *Seamless) GetBalance(ctx context.Context, req *dto.GetBalanceReq) (*dto.GetBalanceResp, error) {
	if err := validate.Req(req); err != nil {
		return nil, rpc.InvalidParamsError
	}

	balance, err := s.seamlessService.Balance(ctx, req.PlayerName, req.Currency)
	if err != nil {
		switch err {
		case service.ErrNotEnoughMoneyCode:
			return nil, &rpc.Error{Code: 1, Message: err.Error()}
		case service.ErrIllegalCurrencyCode:
			return nil, &rpc.Error{Code: 2, Message: err.Error()}
		}
		log.Error(err)
		return nil, &rpc.Error{Code: rpc.ServerErrorCode, Message: "fail get balance"}
	}

	resp := new(dto.GetBalanceResp)

	resp.Balance = balance.Amount
	resp.FreeRoundsLeft = balance.FreeRoundLeft

	return resp, nil
}

func (s *Seamless) WithdrawAndDeposit(ctx context.Context, req *dto.WithdrawAndDepositReq) (*dto.WithdrawAndDepositResp, error) {
	if err := validate.Req(req); err != nil {
		return nil, rpc.InvalidParamsError
	}

	transaction := model.Transaction{
		Withdraw:             req.Withdraw,
		Deposit:              req.Deposit,
		TransactionRef:       req.TransactionRef,
		GameRoundRef:         req.GameRoundRef,
		Source:               req.Source,
		Reason:               req.Reason,
		SessionId:            req.SessionId,
		SessionAlternativeId: req.SessionAlternativeId,
		BonusId:              req.BonusId,
		ChargeFreeRounds:     req.ChargeFreeRounds,
	}

	if req.SpinDetails != nil {
		transaction.SpinDetails = &model.SpinDetails{
			BetType: req.SpinDetails.BetType,
			WinType: req.SpinDetails.WinType,
		}
	}

	newBalance, err := s.seamlessService.Transaction(ctx, req.PlayerName, req.Currency, &transaction)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotEnoughMoneyCode):
			return nil, &rpc.Error{Code: 1, Message: err.Error()}
		case errors.Is(err, service.ErrIllegalCurrencyCode):
			return nil, &rpc.Error{Code: 2, Message: err.Error()}
		case errors.Is(err, service.ErrNegativeDepositCode):
			return nil, &rpc.Error{Code: 3, Message: err.Error()}
		case errors.Is(err, service.ErrNegativeWithdrawalCode):
			return nil, &rpc.Error{Code: 4, Message: err.Error()}
		case errors.Is(err, service.ErrSpendingBudgetExceeded):
			return nil, &rpc.Error{Code: 5, Message: err.Error()}
		case errors.Is(err, service.ErrTransactionRollback):
			return nil, &rpc.Error{Code: 6, Message: err.Error()}
		}
		log.Error(err)
		return nil, &rpc.Error{Code: rpc.ServerErrorCode, Message: "fail transaction"}
	}

	resp := new(dto.WithdrawAndDepositResp)

	resp.NewBalance = newBalance.Amount
	resp.TransactionId = strconv.Itoa(transaction.ID)
	resp.FreeRoundsLeft = newBalance.FreeRoundLeft

	return resp, nil
}

type Empty struct{}

func (s *Seamless) RollbackTransaction(ctx context.Context, req *dto.RollbackTransactionReq) (*Empty, error) {
	if err := validate.Req(req); err != nil {
		return nil, rpc.InvalidParamsError
	}

	now := time.Now()
	transactions := model.Transaction{
		TransactionRef:       req.TransactionRef,
		GameID:               req.GameId,
		GameRoundRef:         req.GameRoundRef,
		SessionId:            req.SessionId,
		SessionAlternativeId: req.SessionAlternativeId,
		IsRollback:           true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	err := s.seamlessService.Rollback(ctx, req.PlayerName, &transactions)
	if err != nil {
		log.Error(err)
		return nil, &rpc.Error{Code: rpc.ServerErrorCode, Message: "fail rollback"}
	}

	return &Empty{}, nil
}
