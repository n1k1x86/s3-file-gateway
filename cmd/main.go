package main

import (
	"context"
	"log"
	"os/signal"
	"s3-file-gateway/config"
	"s3-file-gateway/router"
	"syscall"
	"time"

	"github.com/n1k1x86/libs/http_server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	serverCfg := http_server.NewHTTPServerConfig().WithAddr(cfg.HTTPAddr).WithReadTimeout(time.Second * 10).WithWriteTimeout(time.Second * 10).WithIdleTimeout(time.Second * 10)

	mux := router.InitRouter()

	s := http_server.NewHTTPServer(serverCfg).WithMux(mux)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		errChan <- s.Start()
	}()

	select {
	case <-ctx.Done():
		log.Println("shut down by signal")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
		defer shutdownCancel()

		err := s.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
	case err := <-errChan:
		log.Println("shut down by error from errChan")
		if err != nil {
			log.Fatal(err)
		}
	}
}
