package main

import (
	"log"
	"nanoproxy/pkg/proxy"
	"net/http"
	"os"
)

// ProxyRequestHandler intercepts requests and sends them through the proxy
// func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		proxy.ServeHTTP(w, r)
// 	}
// }

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a reverse proxy and pass the actual upstream here
	// proxy, err := proxy.New("https://httpbin.org/")
	proxy, err := proxy.New("http://httpbin.org/")
	if err != nil {
		panic(err)
	}

	// Route all requests to proxy
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request received: " + r.URL.String())
		proxy.ServeHTTP(w, r)
	})

	log.Println("Starting proxy server on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
