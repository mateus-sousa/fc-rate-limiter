package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
	"github.com/mateus-sousa/fc-rate-limiter/internal/config"
	"github.com/mateus-sousa/fc-rate-limiter/internal/repository"
	"net/http"
	"time"
)

var limiterConfigRepository repository.LimiterConfigRepository

var cfg *config.Conf

func main() {
	var err error
	cfg, err = config.LoadConfig(".")
	if err != nil {
		return
	}
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	ping, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
		return
	}
	fmt.Println(ping)
	fmt.Println(client)
	limiterConfigRepository = repository.NewLimiterConfigCacheRepository(client)
	r := chi.NewRouter()
	r.Get("/hello-world", helloWorld)
	fmt.Println("listening in port :8080")
	http.ListenAndServe(":8080", r)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {

	now := time.Now()
	var requestRuleKey string
	requestRuleKey = readUserIP(r)
	ruleTime := cfg.ReqPerSecondsIP
	token := r.Header.Get("API_KEY")
	if token != "" {
		requestRuleKey = token
		ruleTime = cfg.ReqPerSecondsToken
	}
	val, err := limiterConfigRepository.GetRequestsBy(r.Context(), requestRuleKey)
	if err != nil && err.Error() != "redis: nil" {
		panic(err)
	}
	fmt.Println("val")
	fmt.Println(val)
	limiterConfig := repository.LimiterConfig{
		FirstRequestTime:      now,
		AmountRequestInSecond: 1,
	}
	if val != "" {
		fmt.Println(val)
		var storedLimiterConfig repository.LimiterConfig
		err = json.Unmarshal([]byte(val), &storedLimiterConfig)
		if err != nil {
			panic(err)
		}
		if storedLimiterConfig.AmountRequestInSecond >= ruleTime {
			fmt.Println(http.StatusTooManyRequests)
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		reqDiff := now.Sub(storedLimiterConfig.FirstRequestTime)
		if reqDiff < time.Minute {
			limiterConfig.FirstRequestTime = storedLimiterConfig.FirstRequestTime
			limiterConfig.AmountRequestInSecond = storedLimiterConfig.AmountRequestInSecond + 1
		}
	}
	limiterConfigJson, err := json.Marshal(&limiterConfig)
	if err != nil {
		panic(err)
	}
	err = limiterConfigRepository.SetRequestsAmount(r.Context(), requestRuleKey, limiterConfigJson)
	if err != nil {
		panic(err)
	}
	fmt.Println(http.StatusOK)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(requestRuleKey))
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
