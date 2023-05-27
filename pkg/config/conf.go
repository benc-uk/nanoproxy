package config

import (
	"flag"
	"log"
	"os"

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

var configPath = "./config.yaml"

func init() {
	// Config file path can be set with -c or --config or CONF_FILE env var

	flag.StringVar(&configPath, "config", "./config.yaml", "Path to config file")
	flag.StringVar(&configPath, "c", "./config.yaml", "Path to config file")
	flag.Parse()

	confPathEnv := os.Getenv("CONF_FILE")
	if confPathEnv != "" {
		configPath = confPathEnv
	}
}

func Load() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Config error: %v", err)
		return nil, err
	}

	log.Println("Config loaded from: " + configPath)

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

// Write the config to a file
func (c Config) Write() error {
	d, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, d, 0644)
	if err != nil {
		return err
	}

	return nil
}
