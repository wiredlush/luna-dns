package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Host struct {
	Host string `yaml:"host"`
	IP   string `yaml:"ip"`
}

type DNS struct {
	Addr    string `yaml:"addr"`
	Network string `yaml:"network"`
}

type Config struct {
	Addr       string   `yaml:"addr"`
	Network    string   `yaml:"network"`
	LogFile    string   `yaml:"log_file"`
	DNS        []DNS    `yaml:"dns"`
	Hosts      []Host   `yaml:"hosts"`
	Blocklists []string `yaml:"blocklists"`
	CacheTTL   int64    `yaml:"cache_ttl"`
}

func Load(filepath string) (*Config, error) {
	confBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(confBytes, config)

	return config, err
}
