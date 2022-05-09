package main

import (
	"context"
	"github.com/AlLevykin/cutwell/internal/api/handler"
	"github.com/AlLevykin/cutwell/internal/api/server"
	"github.com/AlLevykin/cutwell/internal/app/store"
	"github.com/caarlos0/env/v6"
	"os"
	"os/signal"
	"sync"
	"time"
)

type config struct {
	Addr            string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL"  envDefault:"127.0.0.1:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func ServeApp(ctx context.Context, wg *sync.WaitGroup, srv *server.Server) {
	defer wg.Done()
	srv.Start()
	<-ctx.Done()
	srv.Stop()
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	ls := store.NewLinkStore(
		store.Config{
			KeyLength: 9,
			BaseURL:   cfg.BaseURL,
		},
		cfg.FileStoragePath)
	defer ls.Save()
	r := handler.NewRouter(ls)
	srv := server.NewServer(
		server.Config{
			Addr:              cfg.Addr,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			ReadHeaderTimeout: 30 * time.Second,
			CancelTimeout:     2 * time.Second,
		},
		r)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go ServeApp(ctx, wg, srv)

	<-ctx.Done()
	cancel()
	wg.Wait()
}
