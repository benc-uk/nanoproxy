package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	proxyName    = "Nanoproxy"
	proxyVersion = "0.0.1"
)

// New takes target host URL and creates a reverse proxy
func New(targetURL string) (*httputil.ReverseProxy, error) {
	log.Printf("Creating proxy with upstream URL: %v\n", targetURL)
	incomingUrl, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(incomingUrl)

	proxy.Director = nil
	proxy.Rewrite = modifyRequest(incomingUrl)
	proxy.ModifyResponse = modifyResponse()

	return proxy, nil
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		resp.Header.Set("X-Proxy", proxyName+"/"+proxyVersion)
		return nil
	}
}

func modifyRequest(url *url.URL) func(*httputil.ProxyRequest) {
	return func(proxyReq *httputil.ProxyRequest) {
		proxyReq.SetXForwarded()
		proxyReq.SetURL(url)
	}
}
