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

// Builds a httputil.ReverseProxy based on a target URL and timeout
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

	// Hook in our own request/response modifiers
	proxy.Director = nil
	proxy.Rewrite = modifyRequest(incomingURL)
	proxy.ModifyResponse = modifyResponse()

	return proxy, nil
}

// This isn't really doing a lot but could be used to modify the response
func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		resp.Header.Set("X-Proxy", proxyName+"/"+proxyVersion)
		return nil
	}
}

// Setup the request to be sent to the upstream server
func modifyRequest(url *url.URL) func(*httputil.ProxyRequest) {
	return func(proxyReq *httputil.ProxyRequest) {
		// Setting X-Forwarded-For and X-Forwarded-Host headers seems polite
		proxyReq.SetXForwarded()

		// Set the URL to the upstream server
		proxyReq.SetURL(url)

		// IMPORTANT: Preserve the original host header
		proxyReq.Out.Host = proxyReq.In.Host
	}
}
