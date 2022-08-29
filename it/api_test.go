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
	s.getBalance()
	s.transaction()
	s.rollback()
}

func (s *apiTestSuite) getBalance() {
	body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","method":"getBalance","params":{"callerId":1,"playerName":"player1","currency":"EUR","gameId":"riot"},"id":0}`))
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

	s.Equal(`{"jsonrpc":"2.0","result":{"balance":10000},"id":0}`, strings.Trim(string(byteBody), "\n"))
}

func (s *apiTestSuite) transaction() {
	body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","method":"withdrawAndDeposit","params":{"callerId":1,"playerName":"player1","withdraw":400,"deposit":200,"currency":"EUR","transactionRef":"1:UOwGgNHPgq3OkqRE","gameRoundRef":"1wawxl:39","gameId":"riot","reason":"GAME_PLAY_FINAL","sessionId":"qx9sgvvpihtrlug","spinDetails":{"betType":"spin","winType":"standart"}},"id":0}`))
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

	s.Equal(`{"jsonrpc":"2.0","result":{"newBalance":9800,"transactionId":"1"},"id":0}`, strings.Trim(string(byteBody), "\n"))
}

func (s *apiTestSuite) rollback() {
	body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","method":"rollbackTransaction","params":{"callerId":1,"playerName":"player1","transactionRef":"1:UOwGgNHPgq3OkqRE","gameId":"riot","sessionId":"qx9sgvvpihtrlug","gameRoundRef":"1wawxl:39"},"id":0}`))
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

	s.Equal(`{"jsonrpc":"2.0","result":{},"id":0}`, strings.Trim(string(byteBody), "\n"))
}
