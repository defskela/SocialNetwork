package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/service"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			Address: ":8081",
			Timeout: time.Second,
		},
	}

	mux := http.NewServeMux()

	srv := NewServer(cfg, mux)
	assert.NotNil(t, srv)
	assert.Equal(t, ":8081", srv.httpServer.Addr)
}

func TestHandler_Init(t *testing.T) {
	services := &service.Service{}
	h := NewHandler(services)

	router := h.Init()
	assert.NotNil(t, router)
}
