package tests

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestRateLimiterWithIP(t *testing.T) {
	client := http.Client{}
	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", "http://localhost:8080/hello-world", nil)
		if err != nil {
			assert.Nil(t, err)
		}
		res, err := client.Do(req)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, res.StatusCode, http.StatusOK)
	}
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest("GET", "http://localhost:8080/hello-world", nil)
		if err != nil {
			assert.Nil(t, err)
		}
		res, err := client.Do(req)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, res.StatusCode, http.StatusTooManyRequests)
	}
	time.Sleep(time.Second)

	req, err := http.NewRequest("GET", "http://localhost:8080/hello-world", nil)
	if err != nil {
		assert.Nil(t, err)
	}
	res, err := client.Do(req)
	if err != nil {
		assert.Nil(t, err)
	}
	assert.Equal(t, res.StatusCode, http.StatusOK)
}

func TestRateLimiterWithToken(t *testing.T) {
	client := http.Client{}
	for i := 0; i < 100; i++ {
		req, err := http.NewRequest("GET", "http://localhost:8080/hello-world", nil)
		if err != nil {
			assert.Nil(t, err)
		}
		req.Header.Add("API_KEY", "abc123")
		res, err := client.Do(req)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, res.StatusCode, http.StatusOK)
	}
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest("GET", "http://localhost:8080/hello-world", nil)
		if err != nil {
			assert.Nil(t, err)
		}
		req.Header.Add("API_KEY", "abc123")
		res, err := client.Do(req)
		if err != nil {
			assert.Nil(t, err)
		}
		assert.Equal(t, res.StatusCode, http.StatusTooManyRequests)
	}
	time.Sleep(time.Second)

	req, err := http.NewRequest("GET", "http://localhost:8080/hello-world", nil)
	if err != nil {
		assert.Nil(t, err)
	}
	req.Header.Add("API_KEY", "abc123")
	res, err := client.Do(req)
	if err != nil {
		assert.Nil(t, err)
	}
	assert.Equal(t, res.StatusCode, http.StatusOK)
}
