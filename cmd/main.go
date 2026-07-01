package main

import (
	"context"
	"log"
	"os/signal"
	"s3-file-gateway/router"
	"syscall"
	"time"

	"github.com/n1k1x86/libs/http_server"
)

func main() {
	cfg := http_server.NewHTTPServerConfig().WithAddr("localhost:8080").WithReadTimeout(time.Second * 10).WithWriteTimeout(time.Second * 10).WithIdleTimeout(time.Second * 10)

	mux := router.InitRouter()

	s := http_server.NewHTTPServer(cfg).WithMux(mux)

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
