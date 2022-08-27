package main

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"seamless-api-wrapper/config"
	"seamless-api-wrapper/internal/logger"
	"seamless-api-wrapper/internal/rpc"
	"seamless-api-wrapper/internal/transport"
	"syscall"
)

func main() {
	configFile := flag.String("config", "./config.toml", "config file")

	flag.Parse()

	logger.InitLogger(os.Stderr, log.InfoLevel)

	cfg, err := config.ParseServerConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	serverConf := cfg.Server

	httpTransport := transport.NewHttpTransport(serverConf.Address, serverConf.ReadTimeout.Duration, serverConf.WriteTimeout.Duration)

	rpcServer := rpc.NewServer(httpTransport)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := rpcServer.Run(ctx); err != nil {
			log.Error(err)
		}
	}()
	log.Info("Server Started")
	<-done
	log.Info("Server Stopped")

	cancel()

}
