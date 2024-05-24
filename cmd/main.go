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
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var limiterConfigRepository repository.LimiterConfigRepository

var cfg *config.Conf
var m sync.Mutex

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
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
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
		return
	}
	limiterConfigRepository = repository.NewLimiterConfigCacheRepository(client)
	m = sync.Mutex{}
	r := chi.NewRouter()
	r.Get("/hello-world", helloWorld)
	fmt.Println("listening in port :8080")
	fmt.Println(cfg.ReqPerSecondsIP)
	fmt.Println(cfg.BlockedTimeIP)
	go func() {
		http.ListenAndServe(":8080", r)
	}()
	<-sigCh
	client.FlushDB(context.Background())
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	var requestRuleKey string
	requestRuleKey = readUserIP(r)
	ruleTime := cfg.ReqPerSecondsIP
	token := r.Header.Get("API_KEY")
	blockedTime := time.Duration(cfg.BlockedTimeIP) * time.Second
	if token != "" {
		requestRuleKey = token
		ruleTime = cfg.ReqPerSecondsToken
		blockedTime = time.Duration(cfg.BlockedTimeToken) * time.Second
	}
	m.Lock()
	defer m.Unlock()
	val, err := limiterConfigRepository.GetRequestsBy(r.Context(), requestRuleKey)
	if err != nil && err.Error() != "redis: nil" {
		panic(err)
	}
	if val == "" {
		limiterConfig := repository.LimiterConfig{
			FirstRequestTime:      now,
			AmountRequestInSecond: 1,
			Blocked:               false,
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
		return
	}
	var storedLimiterConfig repository.LimiterConfig
	err = json.Unmarshal([]byte(val), &storedLimiterConfig)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	reqDiff := now.Sub(storedLimiterConfig.FirstRequestTime)
	if storedLimiterConfig.AmountRequestInSecond < ruleTime && reqDiff < time.Second {
		limiterConfig := repository.LimiterConfig{
			FirstRequestTime:      storedLimiterConfig.FirstRequestTime,
			AmountRequestInSecond: storedLimiterConfig.AmountRequestInSecond + 1,
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
		return
	}
	blockedDiff := now.Sub(storedLimiterConfig.BlockedTime)
	if storedLimiterConfig.AmountRequestInSecond >= ruleTime && reqDiff < time.Second || (storedLimiterConfig.Blocked == true && blockedDiff < blockedTime) {
		limiterConfig := repository.LimiterConfig{
			FirstRequestTime:      storedLimiterConfig.FirstRequestTime,
			AmountRequestInSecond: storedLimiterConfig.AmountRequestInSecond + 1,
			Blocked:               true,
			BlockedTime:           now,
		}
		if storedLimiterConfig.Blocked == true {
			limiterConfig.BlockedTime = storedLimiterConfig.BlockedTime
		}
		limiterConfigJson, err := json.Marshal(&limiterConfig)
		if err != nil {
			panic(err)
		}
		err = limiterConfigRepository.SetRequestsAmount(r.Context(), requestRuleKey, limiterConfigJson)
		if err != nil {
			panic(err)
		}
		fmt.Println(http.StatusTooManyRequests)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	if reqDiff >= time.Second && (storedLimiterConfig.Blocked == false || (storedLimiterConfig.Blocked == true && blockedDiff > blockedTime)) {
		limiterConfig := repository.LimiterConfig{
			FirstRequestTime:      now,
			AmountRequestInSecond: 1,
			Blocked:               false,
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
		return
	}
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	auxIP := strings.Split(IPAddress, ":")
	return auxIP[0]
}
