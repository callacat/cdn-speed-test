package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 定义了应用的配置结构
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

// LoadConfig 从指定路径加载配置文件
func LoadConfig(path string) (*Config, error) {
	// 设置默认值
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


	// 如果文件存在则读取
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(data, cfg)
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
        // 如果文件存在但有其他错误
        return nil, err
    }
    // 如果文件不存在，则使用默认配置

	return cfg, nil
}
