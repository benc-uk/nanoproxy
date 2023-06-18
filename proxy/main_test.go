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

var simpleConf = config.Config{
	Rules: []config.Rule{
		{
			Path:      "/api",
			Upstream:  "example",
			StripPath: true,
		},
		{
			Path:      "/nostrip",
			Upstream:  "example",
			StripPath: false,
		},
		{
			Path:     "/badupstream",
			Upstream: "notexist",
		},
	},
	Upstreams: []config.Upstream{
		{
			Name:   "example",
			Host:   "example.net",
			Scheme: "https",
		},
		{
			Name: "notexist",
			Host: "notexist.zzzyyy.dummy",
		},
	},
}

func TestMain(m *testing.M) {
	os.Setenv("DEBUG", "1")

	nanoProxy.addRoutes()

	m.Run()
}

func TestNoConfRoot404(t *testing.T) {
	// No config applied
	nanoProxy.applyConfig(nil, timeout)

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

func TestProxyPath200(t *testing.T) {
	nanoProxy.applyConfig(&simpleConf, timeout)

	request, _ := http.NewRequest(http.MethodGet, "/api", nil)
	response := httptest.NewRecorder()

	nanoProxy.mainHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", response.Code)
	}
}

func TestProxyPath404(t *testing.T) {
	nanoProxy.applyConfig(&simpleConf, timeout)

	request, _ := http.NewRequest(http.MethodGet, "/cake", nil)
	response := httptest.NewRecorder()

	nanoProxy.mainHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", response.Code)
	}
}

func TestProxyPathNoStrip404(t *testing.T) {
	nanoProxy.applyConfig(&simpleConf, timeout)

	request, _ := http.NewRequest(http.MethodGet, "/nostrip", nil)
	response := httptest.NewRecorder()

	nanoProxy.mainHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", response.Code)
	}
}

func TestProxyPathBadUpstream502(t *testing.T) {
	nanoProxy.applyConfig(&simpleConf, timeout)

	request, _ := http.NewRequest(http.MethodGet, "/badupstream", nil)
	response := httptest.NewRecorder()

	nanoProxy.mainHandler(response, request)

	if response.Code != http.StatusBadGateway {
		t.Errorf("Expected 502, got %d", response.Code)
	}
}
