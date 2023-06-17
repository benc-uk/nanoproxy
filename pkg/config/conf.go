package config

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the main configuration for the proxy
type Config struct {
	Upstreams []Upstream `yaml:"upstreams"`
	Rules     []Rule     `yaml:"rules"`
	Filepath  string     `yaml:"-"`
}

// Upstream is a backend server
type Upstream struct {
	Name   string `yaml:"name"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Scheme string `yaml:"scheme"`
}

// Rule sets host and/or path to match and the upstream to use
type Rule struct {
	Path      string `yaml:"path"`
	Upstream  string `yaml:"upstream"`
	MatchMode string `yaml:"matchMode"`
	Host      string `yaml:"host"`
	StripPath bool   `yaml:"stripPath"`
}

var configPath = "./config.yaml"

// Setup handles the command line arguments and CONF_FILE env var
// Sets the global configPath variable
func Setup() {
	log.Println("************************** Setting up config")
	// Config file path can be set with -c or --config or CONF_FILE env var
	flag.StringVar(&configPath, "config", "./config.yaml", "Path to config file")
	flag.StringVar(&configPath, "c", "./config.yaml", "Path to config file")
	flag.Parse()

	confPathEnv := os.Getenv("CONF_FILE")
	if confPathEnv != "" {
		configPath = confPathEnv
	}
}

// Simple getter for configPath
func GetPath() string {
	return configPath
}

// Load reads the configuration file and returns the configuration.
// It returns an error if the configuration file cannot be loaded.
func Load() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Config error: %v", err)

		return nil, err
	}

	log.Println("Loading config from: " + configPath)

	// Load config file
	conf := Config{}

	err = yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		log.Printf("Config error: %v", err)
		return nil, err
	}

	if os.Getenv("DEBUG") != "" {
		log.Printf("Config dump:\n %+v\n", conf)
	}

	return &conf, nil
}

// Write the config to the config file
func (c Config) Write() error {
	d, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, d, 0600)
	if err != nil {
		return err
	}

	return nil
}

// Dump the config to a string
func (c Config) Dump() string {
	d, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}

	return string(d)
}
