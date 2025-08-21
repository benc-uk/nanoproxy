// ----------------------------------------------------------------------------
// Copyright (c) Ben Coleman, 2023. Licensed under the MIT License.
// NanoProxy reverse proxy server
// ----------------------------------------------------------------------------

package main

import (
	b64 "encoding/base64"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/benc-uk/nanoproxy/pkg/config"
	"github.com/fsnotify/fsnotify"
)

var version = "0.0.0"

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

	if os.Getenv("CONFIG_B64") != "" {
		confBase64 := os.Getenv("CONFIG_B64")

		log.Printf("Config provided from env variable")

		// Decode as Base64
		confBytes, err := b64.StdEncoding.DecodeString(confBase64)
		if err != nil {
			log.Fatalf("Error decoding base64 config: %v", err)
		}

		// Write decoded config to the config file
		err = os.WriteFile(config.GetPath(), confBytes, 0600)
		if err != nil {
			log.Fatalf("Error writing config file: %v", err)
		}

		log.Printf("Env var config file written to: %s", config.GetPath())
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
						log.Println("Warning: config file not loaded, proxy will do nothing")
					}

					// Update & process new config
					nanoProxy.applyConfig(configData, timeout)
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

	// Load config from file
	configData, err := config.Load()
	if err != nil {
		log.Println("Warning: config file not loaded, proxy will do nothing")
	}

	nanoProxy.applyConfig(configData, timeout)
	nanoProxy.startServer(port, timeout, certPath)
}
