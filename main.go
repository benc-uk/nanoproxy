package main

import (
	"log"
	"nanoproxy/pkg/proxy"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"

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
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	c := Config{}
	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	log.Printf("config: %+v", c)

	proxies := make(map[string]*httputil.ReverseProxy)
	for _, u := range c.Upstreams {
		p, err := proxy.New(u.Scheme + "://" + u.Host + ":" + strconv.Itoa(u.Port))

		if err != nil {
			log.Fatalf("proxy error: %v", err)
			continue
		}

		proxies[u.Name] = p
	}

	// All requests flow through this handler and are routed to the correct upstream
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request received: " + r.URL.String())

		// Find matching rule
		for _, rule := range c.Rules {
			matched := false

			if rule.MatchMode != "prefix" && rule.MatchMode != "exact" {
				log.Printf("Invalid match mode for rule host:%s path:%s", rule.Host, rule.Path)
				continue
			}

			// match on prefix
			if rule.MatchMode == "prefix" && strings.HasPrefix(r.URL.Path, rule.Path) {
				matched = true
			}
			if rule.MatchMode == "exact" && r.URL.Path == rule.Path {
				matched = true
			}

			if matched {
				// Find matching upstream
				p := proxies[rule.Upstream]
				if p == nil {
					log.Printf("Upstream '%s' not found", rule.Upstream)
					continue
				}

				p.ServeHTTP(w, r)

				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No matching rule for host & path"))

	})

	log.Println("Starting proxy server on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
