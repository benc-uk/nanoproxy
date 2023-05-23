package main

import (
	"flag"
	"log"
	"nanoproxy/pkg/proxy"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Upstreams []Upstream `yaml:"upstreams"`
	Rules     []Rule     `yaml:"rules"`
}

type Upstream struct {
	Name   string `yaml:"name"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Scheme string `yaml:"scheme"`
}

type Rule struct {
	Path      string `yaml:"path"`
	Upstream  string `yaml:"upstream"`
	MatchMode string `yaml:"matchMode"`
	Host      string `yaml:"host"`
	StripPath bool   `yaml:"stripPath"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	log.Println("Config loaded from: " + *configPath)

	c := Config{}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	if os.Getenv("DEBUG") != "" {
		log.Printf("Config dump:\n %+v\n", c)
	}

	proxies := make(map[string]*httputil.ReverseProxy)

	for _, u := range c.Upstreams {
		scheme := u.Scheme
		if scheme == "" {
			scheme = "http"
		}

		if u.Port == 0 {
			u.Port = 80
		}

		proxy, err := proxy.New(scheme + "://" + u.Host + ":" + strconv.Itoa(u.Port))

		if err != nil {
			log.Fatalf("proxy error: %v", err)
			continue
		}

		proxies[u.Name] = proxy
	}

	// All requests flow through this handler and are routed to the correct upstream
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEBUG") != "" {
			log.Println("Request received: " + r.URL.String())
		}

		// TODO: Optimise this for high volumes of requests and rules

		// Find matching rule
		for _, rule := range c.Rules {
			matched := false

			if !(rule.MatchMode == "" || rule.MatchMode == "prefix" || rule.MatchMode == "exact") {
				log.Printf("Invalid match mode found: %s", rule.MatchMode)
				continue
			}

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
				p := proxies[rule.Upstream]
				if p == nil {
					log.Printf("Upstream '%s' not found in config", rule.Upstream)
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
	})

	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Println("Starting proxy server on port: " + port)

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
