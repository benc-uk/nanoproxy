package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/benc-uk/nanoproxy/pkg/config"
)

var nanoProxy = &NanoProxy{}
var timeout = 5 * time.Second

func TestMain(m *testing.M) {
	os.Setenv("DEBUG", "1")

	nanoProxy.addRoutes()

	m.Run()
}

func TestNoConfRoot404(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	nanoProxy.mainHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", response.Code)
	}
}

func TestDebugConfDump200(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/.nanoproxy/config", nil)
	response := httptest.NewRecorder()

	nanoProxy.mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", response.Code)
	}
}

func TestProxyPathRule200(t *testing.T) {
	conf := config.Config{
		Rules: []config.Rule{
			{
				Path:      "/api",
				Upstream:  "example",
				StripPath: true,
			},
		},
		Upstreams: []config.Upstream{
			{
				Name:   "example",
				Host:   "example.net",
				Scheme: "https",
			},
		},
	}

	nanoProxy.processConfig(conf, timeout)

	request, _ := http.NewRequest(http.MethodGet, "/api", nil)
	response := httptest.NewRecorder()

	nanoProxy.mainHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", response.Code)
	}
}
