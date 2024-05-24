package main

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mateus-sousa/fc-rate-limiter/internal/config"
	"github.com/mateus-sousa/fc-rate-limiter/internal/infra"
	"github.com/mateus-sousa/fc-rate-limiter/internal/repository"
	"github.com/mateus-sousa/fc-rate-limiter/pkg"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var limiterConfigRepository repository.LimiterConfigRepository

var cfg *config.Conf

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	var err error
	cfg, err = config.LoadConfig(".")
	if err != nil {
		panic(err)
		return
	}
	ctx := context.Background()
	client, err := infra.GetRedisClient(ctx)
	if err != nil {
		panic(err)
		return
	}
	limiterConfigRepository = repository.NewLimiterConfigCacheRepository(client)
	rateLimiter := pkg.NewRateLimiter(limiterConfigRepository, cfg)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(rateLimiter.RateLimiterMiddleware)
	r.Get("/hello-world", helloWorld)
	go func() {
		http.ListenAndServe(":8080", r)
	}()
	<-sigCh
	client.FlushDB(context.Background())
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World"))
}
