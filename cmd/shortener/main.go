package main

import (
	"context"
	"github.com/AlLevykin/cutwell/internal/api/handler"
	"github.com/AlLevykin/cutwell/internal/api/server"
	"github.com/AlLevykin/cutwell/internal/app/store"
	"os"
	"os/signal"
	"sync"
	"time"
)

func ServeApp(ctx context.Context, wg *sync.WaitGroup, srv *server.Server) {
	defer wg.Done()
	srv.Start()
	<-ctx.Done()
	srv.Stop()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	ls := store.NewLinkStore(9)
	r := handler.NewRouter(ls)
	cfg := server.Config{
		Addr:              ":8080",
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		CancelTimeout:     2 * time.Second,
	}
	srv := server.NewServer(cfg, r)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go ServeApp(ctx, wg, srv)

	<-ctx.Done()
	cancel()
	wg.Wait()
}
