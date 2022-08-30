package it

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"os"
	"seamless-api-wrapper/internal/api/seamless"
	"seamless-api-wrapper/internal/config"
	"seamless-api-wrapper/internal/logger"
	"seamless-api-wrapper/internal/postgres"
	"seamless-api-wrapper/internal/rpc"
	"seamless-api-wrapper/internal/transport"
	"strings"
	"syscall"
	"testing"
)

var ctx, cancel = context.WithCancel(context.Background())

type apiTestSuite struct {
	suite.Suite
	addr   string
	dbName string
	dbConn *sqlx.DB
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, &apiTestSuite{})
}

func (s *apiTestSuite) SetupSuite() {
	cfg, err := config.ParseServerConfig("test_config.toml")
	s.Require().NoError(err)

	s.addr = cfg.Server.Address
	s.dbName = cfg.Postgres.DB

	logger.InitLogger(os.Stderr, log.InfoLevel)

	serverConf := cfg.Server

	httpTransport := transport.NewHttpTransport(serverConf.Address, serverConf.ReadTimeout.Duration, serverConf.WriteTimeout.Duration)

	rpcServer := rpc.NewServer(httpTransport)

	db, err := postgres.InitDB(&cfg.Postgres)
	s.Require().NoError(err)

	s.dbConn = db

	seamlessService := postgres.NewSeamlessService(db)

	api := seamless.NewSeamless(seamlessService)

	rpcServer.Register("getBalance", rpc.HandlerWithPointer(api.GetBalance))
	rpcServer.Register("withdrawAndDeposit", rpc.HandlerWithPointer(api.WithdrawAndDeposit))
	rpcServer.Register("rollbackTransaction", rpc.HandlerWithPointer(api.RollbackTransaction))

	go func() {
		if err := rpcServer.Run(ctx); err != nil {
			log.Error(err)
		}
	}()
}

func (s *apiTestSuite) TearDownSuite() {
	err := s.dbConn.Close()
	s.NoError(err)
	cancel()

	p, _ := os.FindProcess(syscall.Getpid())
	_ = p.Signal(syscall.SIGINT)
}

func (s *apiTestSuite) Test_Seamless() {
	getBalanceReq := `{"jsonrpc":"2.0","method":"getBalance","params":{"callerId":1,"playerName":"player1","currency":"EUR","gameId":"riot"},"id":0}`
	getBalanceResp := `{"jsonrpc":"2.0","result":{"balance":10000},"id":0}`
	s.senMessage(getBalanceReq, getBalanceResp)

	transactionReq := `{"jsonrpc":"2.0","method":"withdrawAndDeposit","params":{"callerId":1,"playerName":"player1","withdraw":400,"deposit":200,"currency":"EUR","transactionRef":"1:UOwGgNHPgq3OkqRE","gameRoundRef":"1wawxl:39","gameId":"riot","reason":"GAME_PLAY_FINAL","sessionId":"qx9sgvvpihtrlug","spinDetails":{"betType":"spin","winType":"standart"}},"id":0}`
	transactionResp := `{"jsonrpc":"2.0","result":{"newBalance":9800,"transactionId":"1"},"id":0}`
	s.senMessage(transactionReq, transactionResp)

	rollbackReq := `{"jsonrpc":"2.0","method":"rollbackTransaction","params":{"callerId":1,"playerName":"player1","transactionRef":"1:UOwGgNHPgq3OkqRE","gameId":"riot","sessionId":"qx9sgvvpihtrlug","gameRoundRef":"1wawxl:39"},"id":0}`
	rollbackResp := `{"jsonrpc":"2.0","result":{},"id":0}`
	s.senMessage(rollbackReq, rollbackResp)

	getBalanceReq = `{"jsonrpc":"2.0","method":"getBalance","params":{"callerId":1,"playerName":"player1","currency":"EUR","gameId":"riot"},"id":0}`
	getBalanceResp = `{"jsonrpc":"2.0","result":{"balance":10000,"freeroundsLeft":0},"id":0}`
	s.senMessage(getBalanceReq, getBalanceResp)
}

func (s *apiTestSuite) Test_SeamlessRollback() {
	_, err := s.dbConn.Exec(`INSERT INTO balances(player_name, currency_id, amount, game_id, created_at, updated_at) VALUES ('player2', 1, 20000, 'riot', NOW(), NOW())`)
	s.NoError(err)

	getBalanceReq := `{"jsonrpc":"2.0","method":"getBalance","params":{"callerId":1,"playerName":"player2","currency":"EUR","gameId":"riot"},"id":0}`
	getBalanceResp := `{"jsonrpc":"2.0","result":{"balance":20000},"id":0}`
	s.senMessage(getBalanceReq, getBalanceResp)

	rollbackReq := `{"jsonrpc":"2.0","method":"rollbackTransaction","params":{"callerId":1,"playerName":"player2","transactionRef":"2:UOwGgNHPgq3OkqRE","gameId":"riot","sessionId":"qx9sgvvpihtrlug","gameRoundRef":"1wawxl:39"},"id":0}`
	rollbackResp := `{"jsonrpc":"2.0","result":{},"id":0}`
	s.senMessage(rollbackReq, rollbackResp)

	transactionReq := `{"jsonrpc":"2.0","method":"withdrawAndDeposit","params":{"callerId":1,"playerName":"player2","withdraw":100,"deposit":50,"currency":"EUR","transactionRef":"2:UOwGgNHPgq3OkqRE","gameRoundRef":"1wawxl:39","gameId":"riot","reason":"GAME_PLAY_FINAL","sessionId":"qx9sgvvpihtrlug","spinDetails":{"betType":"spin","winType":"standart"}},"id":0}`
	transactionResp := `{"jsonrpc":"2.0","error":{"code":6,"message":"ErrTransactionRollback"},"id":0}`
	s.senMessage(transactionReq, transactionResp)
}

func (s *apiTestSuite) Test_SeamlessTransactions() {
	_, err := s.dbConn.Exec(`INSERT INTO balances(player_name, currency_id, amount, game_id, created_at, updated_at) VALUES ('player3', 1, 200, 'riot', NOW(), NOW())`)
	s.NoError(err)

	transactionReq := `{"jsonrpc":"2.0","method":"withdrawAndDeposit","params":{"callerId":1,"playerName":"player3","withdraw":100,"deposit":50,"currency":"EUR","transactionRef":"3:UOwGgNHPgq3OkqRE","gameRoundRef":"1wawxl:39","gameId":"riot","reason":"GAME_PLAY_FINAL","sessionId":"qx9sgvvpihtrlug","spinDetails":{"betType":"spin","winType":"standart"}},"id":0}`
	transactionResp := `{"jsonrpc":"2.0","result":{"newBalance":150,"transactionId":"5"},"id":0}`
	s.senMessage(transactionReq, transactionResp)

	transactionReq = `{"jsonrpc":"2.0","method":"withdrawAndDeposit","params":{"callerId":1,"playerName":"player3","withdraw":100,"deposit":50,"currency":"EUR","transactionRef":"3:UOwGgNHPgq3OkqRE","gameRoundRef":"1wawxl:39","gameId":"riot","reason":"GAME_PLAY_FINAL","sessionId":"qx9sgvvpihtrlug","spinDetails":{"betType":"spin","winType":"standart"}},"id":0}`
	transactionResp = `{"jsonrpc":"2.0","result":{"newBalance":150,"transactionId":"5"},"id":0}`
	s.senMessage(transactionReq, transactionResp)

	getBalanceReq := `{"jsonrpc":"2.0","method":"getBalance","params":{"callerId":1,"playerName":"player3","currency":"EUR","gameId":"riot"},"id":0}`
	getBalanceResp := `{"jsonrpc":"2.0","result":{"balance":150},"id":0}`
	s.senMessage(getBalanceReq, getBalanceResp)
}

func (s *apiTestSuite) senMessage(reqBody, respBody string) {
	body := bytes.NewReader([]byte(reqBody))
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s", s.addr), body)
	s.NoError(err)

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(req)
	s.NoError(err)

	s.Equal(http.StatusOK, response.StatusCode)

	byteBody, err := io.ReadAll(response.Body)
	response.Body.Close()
	s.NoError(err)

	s.Equal(respBody, strings.Trim(string(byteBody), "\n"))
}
