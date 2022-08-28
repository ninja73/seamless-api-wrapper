package seamless

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"seamless-api-wrapper/internal/model"
	"seamless-api-wrapper/internal/service"
	"seamless-api-wrapper/package/dto"
	"strconv"
)

type Seamless struct {
	seamlessService service.SeamlessService
}

func NewSeamless(seamlessService service.SeamlessService) *Seamless {
	return &Seamless{seamlessService: seamlessService}
}

func (s *Seamless) GetBalance(ctx context.Context, req *dto.GetBalanceReq) (*dto.GetBalanceResp, error) {
	balance, err := s.seamlessService.Balance(ctx, req.PlayerName, req.Currency)
	if err != nil {
		log.Error(err)
		return nil, errors.New("fail get balance")
	}

	resp := new(dto.GetBalanceResp)

	resp.Balance = balance.Amount
	resp.FreeRoundsLeft = balance.FreeRoundLeft

	return resp, nil
}

func (s *Seamless) WithdrawAndDeposit(ctx context.Context, req *dto.WithdrawAndDepositReq) (*dto.WithdrawAndDepositResp, error) {
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
	newBalance, err := s.seamlessService.Transaction(ctx, req.PlayerName, req.Currency, &transaction)
	if err != nil {

	}

	resp := new(dto.WithdrawAndDepositResp)

	resp.NewBalance = newBalance.Amount
	resp.TransactionId = strconv.Itoa(transaction.ID)
	resp.FreeRoundsLeft = newBalance.FreeRoundLeft

	return resp, nil
}

type Empty struct{}

func (s *Seamless) RollbackTransaction(_ context.Context, _ *dto.RollbackTransactionReq) (*Empty, error) {
	return &Empty{}, nil
}
