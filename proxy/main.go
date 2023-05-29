// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoProxy reverse proxy server
// ----------------------------------------------------------------------------

package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/benc-uk/nanoproxy/pkg/config"
	"github.com/fsnotify/fsnotify"
)

var version = "0.0.0"

type NanoProxy struct {
	proxies map[string]*httputil.ReverseProxy
	config  config.Config
}

func main() {
	log.Printf("Starting NanoProxy version: %s", version)

	port := "8080"
	timeout := 5 * time.Second

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	if os.Getenv("TIMEOUT") != "" {
		t, err := strconv.Atoi(os.Getenv("TIMEOUT"))
		if err != nil {
			log.Fatalf("Invalid timeout value: %s", os.Getenv("TIMEOUT"))
		}

		timeout = time.Duration(t) * time.Second
	}

	nanoProxy := &NanoProxy{}

	// Setup file watcher for config file
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for config file changes
	go func() {
		last := time.Now()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op == fsnotify.Write {
					log.Println("Config file changed:", event.Name)
					// Ignore multiple events in a short time
					if time.Since(last) < 500*time.Millisecond {
						continue
					}

					last = time.Now()

					// Eurgh, see https://github.com/fsnotify/fsnotify/issues/372
					time.Sleep(200 * time.Millisecond)

					// Update & process new config
					nanoProxy.processConfig(timeout)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Println("Config watch error:", err)
			}
		}
	}()

	log.Println("Watching config file: " + config.GetPath())
	err = watcher.Add(config.GetPath())

	if err != nil {
		if os.IsNotExist(err) {
			// Try to create config file and watch it
			// Ignore errors in here otherwise there's no escape
			log.Println("Config file not found, creating empty file and watching")

			_ = os.WriteFile(config.GetPath(), []byte(""), 0600)
			_ = watcher.Add(config.GetPath())
		} else {
			log.Fatal(err)
		}
	}

	// Process loaded config file
	nanoProxy.processConfig(timeout)

	// All requests flow through this main handler
	http.HandleFunc("/", nanoProxy.handle)

	if os.Getenv("DEBUG") != "" {
		log.Println("Debug mode enabled, exposing /.config endpoint")

		http.HandleFunc("/.config", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(nanoProxy.config.Dump()))
		})
	}

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	log.Println("Server listening on port: " + port)

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func (np *NanoProxy) processConfig(timeout time.Duration) {
	// Initial load of config file
	configData, err := config.Load()
	if err != nil {
		log.Println("Warning: no config file, proxy will do nothing")

		configData = &config.Config{}
	}

	np.config = *configData

	np.proxies = make(map[string]*httputil.ReverseProxy)

	for _, u := range np.config.Upstreams {
		scheme := u.Scheme
		if scheme == "" {
			scheme = "http"
		}

		if scheme != "http" && scheme != "https" {
			log.Fatalf("Invalid scheme found: %s", scheme)
			continue
		}

		if u.Port == 0 && scheme == "http" {
			u.Port = 80
		}

		if u.Port == 0 && scheme == "https" {
			u.Port = 443
		}

		proxy, err := NewProxy(scheme+"://"+u.Host+":"+strconv.Itoa(u.Port), timeout)
		if err != nil {
			log.Fatalf("proxy error: %v", err)
			continue
		}

		np.proxies[u.Name] = proxy
	}

	for _, rule := range np.config.Rules {
		if !(rule.MatchMode == "" || rule.MatchMode == "prefix" || rule.MatchMode == "exact") {
			log.Printf("Rule error: invalid match mode: %s", rule.MatchMode)
			continue
		}

		if rule.Path == "" {
			log.Printf("Rule error: path is blank, this rule will never match")
			continue
		}
	}

	if len(np.config.Rules) <= 0 {
		log.Printf("Warning: config contains no rules")
	}

	if len(np.proxies) <= 0 {
		log.Printf("Warning: config contains no upstreams")
	}
}

// The main router for all requests
func (np *NanoProxy) handle(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DEBUG") != "" {
		log.Println("Request received: " + r.URL.String())
	}

	// TODO: Optimise this for high volumes of requests and rules

	// Find matching rule, the main routing
	for _, rule := range np.config.Rules {
		matched := false

		// Strip port from host
		hostname := r.Host
		if strings.Contains(hostname, ":") {
			hostname = strings.Split(hostname, ":")[0]
		}

		if os.Getenv("DEBUG") != "" {
			log.Printf("Checking rule host:%s path:%s - against host:%s path:%s",
				rule.Host, rule.Path, hostname, r.URL.Path)
		}

		// Match on host first, empty host matches all
		if rule.Host == "" || hostname == rule.Host {
			// Match on prefix which is the default if not specified
			if (rule.MatchMode == "prefix" || rule.MatchMode == "") && strings.HasPrefix(r.URL.Path, rule.Path) {
				matched = true
			}

			if rule.MatchMode == "exact" && r.URL.Path == rule.Path {
				matched = true
			}
		}

		if matched {
			if os.Getenv("DEBUG") != "" {
				log.Printf("Matched rule: %s_%s_%s", rule.Upstream, rule.Host, rule.Path)
			}

			// Find proxy named by the rule that was matched
			p := np.proxies[rule.Upstream]
			if p == nil {
				log.Printf("Rule error: upstream '%s' not found", rule.Upstream)
				continue
			}

			// Strip path
			if rule.StripPath {
				r.URL.Path = strings.Replace(r.URL.Path, rule.Path, "", 1)
			}

			p.ServeHTTP(w, r)

			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("No matching rule for host & path"))
}
