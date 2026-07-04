package main

import (
	"context"
	"log"
	"os/signal"
	"s3-file-gateway/config"
	"s3-file-gateway/router"
	"s3-file-gateway/s3_storage"
	"syscall"
	"time"

	"github.com/n1k1x86/libs/http_server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	serverCfg := http_server.NewHTTPServerConfig().WithAddr(cfg.HTTPAddr).WithReadTimeout(time.Second * 10).WithWriteTimeout(time.Second * 10).WithIdleTimeout(time.Second * 10)

	s3Storage, err := s3_storage.NewS3Storage(ctx, cfg.S3Key, cfg.S3Secret, cfg.S3Region, cfg.S3Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	mux := router.InitRouter(s3Storage)

	s := http_server.NewHTTPServer(serverCfg).WithMux(mux)

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
