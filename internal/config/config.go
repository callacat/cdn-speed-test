package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config defines the application's configuration structure.
type Config struct {
	Test struct {
		Concurrency int           `yaml:"concurrency"`
		Retries     int           `yaml:"retries"`
		Timeout     time.Duration `yaml:"timeout"`
		LatencyMax  int           `yaml:"latency_max"`
		TopN        int           `yaml:"top_n"`
	} `yaml:"test"`
	HTTP struct {
		TargetURL        string        `yaml:"target_url"`
		SpeedTestURL     string        `yaml:"speed_test_url"`
		SpeedTestTimeout time.Duration `yaml:"speed_test_timeout"`
	} `yaml:"http"`
	IPSource struct {
		APIURL      string `yaml:"api_url"`
		LocalFiles struct {
			IPv4 string `yaml:"ipv4"`
		} `yaml:"local_files"`
	} `yaml:"ip_source"`
	Output struct {
		Format   string `yaml:"format"`
		CSVPath  string `yaml:"csv_path"`
	} `yaml:"output"`
}

// Load loads the configuration from a given path, applying defaults first.
func Load(path string) (*Config, error) {
	// Start with default values.
	cfg := &Config{}
	cfg.Test.Concurrency = 1000
	cfg.Test.Retries = 4
	cfg.Test.Timeout = 5 * time.Second
	cfg.Test.LatencyMax = 400
	cfg.Test.TopN = 20
	cfg.HTTP.TargetURL = "https://www.google.com/generate_204"
	cfg.HTTP.SpeedTestURL = "https://cachefly.cachefly.net/100mb.test"
	cfg.HTTP.SpeedTestTimeout = 30 * time.Second
	cfg.IPSource.LocalFiles.IPv4 = "./ip.txt"
	cfg.Output.Format = "table"
	cfg.Output.CSVPath = "./result.csv"

	// If the config file exists, read it and override defaults.
	data, err := os.ReadFile(path)
	if err != nil {
		// If the file doesn't exist, that's okay; we'll use defaults.
		// If another error occurred, return it.
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		// File exists, so unmarshal it into the config struct.
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
