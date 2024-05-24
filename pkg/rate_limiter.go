package pkg

import (
	"encoding/json"
	"github.com/mateus-sousa/fc-rate-limiter/internal/config"
	"github.com/mateus-sousa/fc-rate-limiter/internal/repository"
	"net/http"
	"strings"
	"time"
)

type RateLimiter struct {
	limiterConfigRepository repository.LimiterConfigRepository
	cfg                     *config.Conf
}

func NewRateLimiter(limiterConfigRepository repository.LimiterConfigRepository, cfg *config.Conf) *RateLimiter {
	return &RateLimiter{limiterConfigRepository: limiterConfigRepository, cfg: cfg}
}

func (rt *RateLimiter) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		var requestRuleKey string
		requestRuleKey = rt.readUserIP(r)
		ruleTime := rt.cfg.ReqPerSecondsIP
		token := r.Header.Get("API_KEY")
		blockedTime := time.Duration(rt.cfg.BlockedTimeIP) * time.Second
		if token != "" {
			requestRuleKey = token
			ruleTime = rt.cfg.ReqPerSecondsToken
			blockedTime = time.Duration(rt.cfg.BlockedTimeToken) * time.Second
		}
		val, err := rt.limiterConfigRepository.GetRequestsBy(r.Context(), requestRuleKey)
		if err != nil && err.Error() != "redis: nil" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if val == "" {
			limiterConfig := repository.LimiterConfig{
				FirstRequestTime:      now,
				AmountRequestInSecond: 1,
				Blocked:               false,
			}
			limiterConfigJson, err := json.Marshal(&limiterConfig)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = rt.limiterConfigRepository.SetRequestsAmount(r.Context(), requestRuleKey, limiterConfigJson)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
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
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = rt.limiterConfigRepository.SetRequestsAmount(r.Context(), requestRuleKey, limiterConfigJson)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
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
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = rt.limiterConfigRepository.SetRequestsAmount(r.Context(), requestRuleKey, limiterConfigJson)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
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
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = rt.limiterConfigRepository.SetRequestsAmount(r.Context(), requestRuleKey, limiterConfigJson)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rt *RateLimiter) readUserIP(r *http.Request) string {
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
