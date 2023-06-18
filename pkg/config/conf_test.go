package config

// Tests for config package
import (
	"log"
	"os"
	"strings"
	"testing"
)

var yamlConfig = `
upstreams:
  - name: my-server-a
    host: some.hostname.here
    scheme: https
  - name: my-server-b
    host: backend.api.example
    port: 3000

rules:
  - upstream: my-server-b
    path: /api
    stripPath: true
  - upstream: my-server-a
    path: /
    host: proxy.example.net
`

func TestMain(m *testing.M) {
	os.Setenv("CONF_FILE", "/tmp/nanoproxy-config.yaml")
	//Setup()

	m.Run()

	os.Remove(GetPath())
}

// Test the config package
func TestConfigMissing(t *testing.T) {
	conf, err := Load()
	if err == nil {
		t.Errorf("Expected error loading config, got none")
	}

	if conf != nil {
		t.Errorf("Expected nil config, got %+v", conf)
	}
}

func TestConfigEmpty(t *testing.T) {
	os.Remove(GetPath())
	_ = os.WriteFile(GetPath(), []byte(""), 0600)

	conf, err := Load()
	if err != nil {
		t.Errorf("Expected no error loading config, got %v", err)
	}

	//nolint:staticcheck
	if conf == nil {
		t.Errorf("Expected non-nil config, got nil")
	}

	//nolint:staticcheck
	if len(conf.Rules) != 0 {
		t.Errorf("Expected empty rules, got %+v", conf.Rules)
	}

	if len(conf.Upstreams) != 0 {
		t.Errorf("Expected empty upstreams, got %+v", conf.Upstreams)
	}
}

func TestConfigInvalid(t *testing.T) {
	_ = os.WriteFile(GetPath(), []byte("invalid"), 0600)

	conf, err := Load()
	if err == nil {
		t.Errorf("Expected error loading config, got none")
	}

	if conf != nil {
		t.Errorf("Expected nil config, got %+v", conf)
	}
}

func TestConfigValid(t *testing.T) {
	_ = os.WriteFile(GetPath(), []byte(yamlConfig), 0600)

	conf, err := Load()
	if err != nil {
		t.Errorf("Expected no error loading config, got %v", err)
	}

	//nolint:staticcheck
	if conf == nil {
		t.Errorf("Expected non-nil config, got nil")
	}

	//nolint:staticcheck
	if len(conf.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %+v", conf.Rules)
	}

	if len(conf.Upstreams) != 2 {
		t.Errorf("Expected 2 upstreams, got %+v", conf.Upstreams)
	}
}

func TestConfigDump(t *testing.T) {
	conf := &Config{
		Rules: []Rule{
			{
				Upstream:  "example",
				Path:      "/api",
				StripPath: true,
			},
		},
	}

	result := conf.Dump()
	if result == "" {
		t.Errorf("Expected non-empty result, got %s", result)
	}

	if strings.Contains(result, "example") == false {
		t.Errorf("Expected example in result, got %s", result)
	}
}

func TestConfigWrite(t *testing.T) {
	conf := &Config{
		Rules: []Rule{
			{
				Upstream:  "example",
				Path:      "/api",
				StripPath: true,
			},
		},
	}

	err := conf.Write()
	if err != nil {
		t.Errorf("Expected no error writing config, got %v", err)
	}

	bytes, err := os.ReadFile(GetPath())
	if err != nil {
		t.Errorf("Expected no error reading config, got %v", err)
	}

	log.Printf("Got config: %s", string(bytes))

	if string(bytes) != conf.Dump() {
		t.Errorf("Expected config to match, got %s", string(bytes))
	}
}
