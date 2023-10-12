package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopkg.in/resty.v1"
)

func TestRun(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	}))
	defer testServer.Close()

	config := &Config{
		URL: testServer.URL,
		Requests: struct {
			Amount    int `mapstructure:"amount"`
			PerSecond int `mapstructure:"per_second"`
		}{
			Amount:    10,
			PerSecond: 2,
		},
	}

	logger, _ := NewLogger()

	client := resty.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})

	go RunWithContext(ctx, config, logger, client, done)

	close(done)
	<-time.After(5 * time.Second)

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Error("The test has timed out")
	}
}
