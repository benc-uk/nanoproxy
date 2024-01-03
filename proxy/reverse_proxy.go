// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoProxy reverse proxies (aka upstreams)
// ----------------------------------------------------------------------------

package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

const (
	proxyName = "Nanoproxy"
)

var hostname string

// Builds a httputil.ReverseProxy based on a target URL and timeout
func NewReverseProxy(targetURL string, timeout time.Duration, hostRewrite bool) (*httputil.ReverseProxy, error) {
	log.Printf("Creating upstream: %v\n", targetURL)

	incomingURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	// This httputil.ReverseProxy is doing a lot of the heavy lifting
	proxy := httputil.NewSingleHostReverseProxy(incomingURL)

	// Check if we should skip TLS verification
	skipTLSVerify := false

	skipTLSVerifyEnv := os.Getenv("TLS_SKIP_VERIFY")
	if skipTLSVerifyEnv != "" {
		skipTLSVerify = true
	}

	// create Transport with timeout
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,

		//nolint:gosec
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
	}

	// Hook in our own request/response modifiers
	proxy.Director = nil
	proxy.Rewrite = modifyRequest(incomingURL, hostRewrite)
	proxy.ModifyResponse = modifyResponse()

	// get hostname of where we are running
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return proxy, nil
}

// This isn't really doing a lot but could be used to modify the response
func modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		// Custom headers to identify the proxy and instance
		resp.Header.Set("X-Proxy", proxyName+"/"+version)
		resp.Header.Set("X-Proxy-Instance", hostname)

		return nil
	}
}

// Setup the request to be sent to the upstream server
func modifyRequest(url *url.URL, hostRewrite bool) func(*httputil.ProxyRequest) {
	return func(proxyReq *httputil.ProxyRequest) {
		// Setting X-Forwarded-For and X-Forwarded-Host headers seems polite
		proxyReq.SetXForwarded()

		// Set the URL to the upstream server
		proxyReq.SetURL(url)

		// IMPORTANT: Preserve the original host header
		if hostRewrite {
			proxyReq.Out.Host = proxyReq.In.Host
		}
	}
}
