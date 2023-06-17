// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoProxy reverse proxy server
// ----------------------------------------------------------------------------

package main

import (
	"crypto/tls"
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
	config  config.Config  // Hold a copy of the config
	mux     *http.ServeMux // Hold a reference to the mux for testing
}

func main() {
	log.Printf("Starting NanoProxy version: %s", version)

	port := "8080"
	timeout := 5 * time.Second
	certPath := ""

	// Used for TLS, directory where cert.pem and key.pem are found
	if os.Getenv("CERT_PATH") != "" {
		certPath = os.Getenv("CERT_PATH")
	}

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
					// Ignore multiple events in a short time
					if time.Since(last) < 500*time.Millisecond {
						continue
					}

					last = time.Now()

					// Eurgh, see https://github.com/fsnotify/fsnotify/issues/372
					time.Sleep(200 * time.Millisecond)

					configData, err := config.Load()
					if err != nil {
						log.Println("Warning: no config file, proxy will do nothing")

						// Create empty config, so we don't panic later
						configData = &config.Config{}
					}

					// Update & process new config
					nanoProxy.processConfig(*configData, timeout)
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
			// Ignore errors in here it's just a best effort
			log.Println("Config file not found, creating empty file and watching")

			_ = os.WriteFile(config.GetPath(), []byte(""), 0600)
			_ = watcher.Add(config.GetPath())
		} else {
			log.Fatal(err)
		}
	}

	// Must be called before any config is loaded
	config.Setup()

	// Load config from file
	configData, err := config.Load()
	if err != nil {
		log.Println("Warning: no config file, proxy will do nothing")

		// Create empty config, so we don't panic later
		configData = &config.Config{}
	}

	nanoProxy.processConfig(*configData, timeout)
	nanoProxy.start(port, timeout, certPath)
}

func (np *NanoProxy) start(port string, timeout time.Duration, certPath string) {
	np.addRoutes()

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		},
		Handler: np.mux,
	}

	useTLS := false
	if certPath != "" {
		log.Printf("Checking cert & key files: %s %s", certPath+"/cert.pem", certPath+"/key.pem")
		useTLS = true

		// Check cert & key files exist
		if _, err := os.Stat(certPath + "/cert.pem"); os.IsNotExist(err) {
			log.Printf("ERROR! Cert file not found: " + certPath + "/cert.pem")
			useTLS = false
		}

		if _, err := os.Stat(certPath + "/key.pem"); os.IsNotExist(err) {
			log.Printf("ERROR! Key file not found: " + certPath + "/key.pem")
			useTLS = false
		}
	}

	if useTLS {
		log.Println("TLS has been enabled, proxy will accept HTTPS traffic on port: " + port)
		err := server.ListenAndServeTLS(certPath+"/cert.pem", certPath+"/key.pem")
		if err != nil {
			panic(err)
		}
	} else {
		log.Println("TLS is disabled, proxy will accept HTTP traffic on port: " + port)
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}
}

func (np *NanoProxy) addRoutes() {
	mux := http.NewServeMux()

	// All requests flow through this main handler
	mux.HandleFunc("/", np.mainHandler)

	if os.Getenv("DEBUG") != "" {
		log.Println("Debug enabled, exposing /.nanoproxy/config endpoint")

		mux.HandleFunc("/.nanoproxy/config", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(np.config.Dump()))
		})
	}

	// Add health check endpoint, weird name to try to avoid clashes
	mux.HandleFunc("/.nanoproxy/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	// Set the mux to our new one
	np.mux = mux
}

// This loads config and creates the reverse proxies
func (np *NanoProxy) processConfig(config config.Config, timeout time.Duration) {
	// This is the map of reverse proxies, keyed by upstream name
	np.proxies = make(map[string]*httputil.ReverseProxy)

	// Construct reverse proxies for each upstream
	// Note the term upstream is used in the config file, but we call them proxies here
	for _, u := range config.Upstreams {
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
			log.Fatalf("Error with reverse proxy: %v", err)
			continue
		}

		np.proxies[u.Name] = proxy
	}

	// Validate & check rules
	for _, rule := range config.Rules {
		if !(rule.MatchMode == "" || rule.MatchMode == "prefix" || rule.MatchMode == "exact") {
			log.Printf("Rule error: invalid match mode: %s", rule.MatchMode)
			continue
		}

		if rule.Path == "" {
			log.Printf("Rule error: path is blank, this rule will never match")
			continue
		}
	}

	if len(config.Rules) <= 0 {
		log.Printf("Warning: config contains no rules")
	}

	if len(np.proxies) <= 0 {
		log.Printf("Warning: config contains no upstreams")
	}

	// Save config
	np.config = config
}

// This is the main router for all proxied requests
func (np *NanoProxy) mainHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DEBUG") != "" {
		log.Println("Request received: " + r.URL.String())
	}

	// TODO: Optimise this for high volumes of requests and rules

	// Find matching rule, the main routing logic
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
			// Match path on prefix which is the default MatchMode
			if (rule.MatchMode == "prefix" || rule.MatchMode == "") && strings.HasPrefix(r.URL.Path, rule.Path) {
				matched = true
			}

			if rule.MatchMode == "exact" && r.URL.Path == rule.Path {
				matched = true
			}
		}

		// Path and/or host was matched to this rule, so proxy the request
		if matched {
			if os.Getenv("DEBUG") != "" {
				log.Printf("Matched rule: %s_%s_%s", rule.Upstream, rule.Host, rule.Path)
			}

			// Find proxy named by the rule that was matched
			proxy := np.proxies[rule.Upstream]
			if proxy == nil {
				log.Printf("Rule error: upstream '%s' not found", rule.Upstream)
				continue
			}

			// Strip path
			if rule.StripPath {
				r.URL.Path = strings.Replace(r.URL.Path, rule.Path, "", 1)
			}

			// It all comes down to this, proxy the request
			proxy.ServeHTTP(w, r)

			// Don't process any more rules
			return
		}
	}

	if os.Getenv("DEBUG") != "" {
		log.Printf("No matching rule for request - host:%s path:%s", r.Host, r.URL.Path)
	}

	// Fall through, no matching rule found so return 404
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("No matching rule for host & path"))
}
