// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoProxy reverse proxies (aka upstreams)
// ----------------------------------------------------------------------------

package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	proxyName    = "Nanoproxy"
	proxyVersion = "0.0.1"
)

// New takes target host URL and creates a reverse proxy
func NewProxy(targetURL string, timeout time.Duration) (*httputil.ReverseProxy, error) {
	log.Printf("Creating upstream: %v\n", targetURL)

	incomingURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	// This httputil.ReverseProxy is doing a lot of the heavy lifting
	proxy := httputil.NewSingleHostReverseProxy(incomingURL)

	// create Transport with timeout
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}

	proxy.Director = nil
	proxy.Rewrite = modifyRequest(incomingURL)
	proxy.ModifyResponse = modifyResponse()

	return proxy, nil
}

func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		// Placeholder for future modifications
		resp.Header.Set("X-Proxy", proxyName+"/"+proxyVersion)
		return nil
	}
}

func modifyRequest(url *url.URL) func(*httputil.ProxyRequest) {
	return func(proxyReq *httputil.ProxyRequest) {
		// Placeholder for future modifications
		proxyReq.SetXForwarded()
		proxyReq.SetURL(url)

		// Preserve the original host header
		proxyReq.Out.Host = proxyReq.In.Host
	}
}
