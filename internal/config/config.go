package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`
	Search  SearchConfig  `mapstructure:"search"`
	Daemon  DaemonConfig  `mapstructure:"daemon"`
	Domain  DomainConfig  `mapstructure:"domain"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type StorageConfig struct {
	Path string `mapstructure:"path"`
}

type SearchConfig struct {
	IndexPath string `mapstructure:"index_path"`
}

type DaemonConfig struct {
	PidFile string `mapstructure:"pid_file"`
	LogFile string `mapstructure:"log_file"`
}

type DomainConfig struct {
	DefaultDomain     string `mapstructure:"default_domain"`
	IsolationMode     string `mapstructure:"isolation_mode"`
	AllowCrossDomain  bool   `mapstructure:"allow_cross_domain"`
	CurrentDomain     string `mapstructure:"current_domain"`
}

func Load(configFile string) (*Config, error) {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("storage.path", "./data/storage")
	viper.SetDefault("search.index_path", "./data/search")
	viper.SetDefault("daemon.pid_file", "/tmp/mcp-mindport.pid")
	viper.SetDefault("daemon.log_file", "/tmp/mcp-mindport.log")
	viper.SetDefault("domain.default_domain", "default")
	viper.SetDefault("domain.isolation_mode", "hierarchical")
	viper.SetDefault("domain.allow_cross_domain", true)
	viper.SetDefault("domain.current_domain", "default")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mcp-mindport")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Ensure directories exist
	if err := os.MkdirAll(config.Storage.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	if err := os.MkdirAll(config.Search.IndexPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create search index directory: %w", err)
	}

	return &config, nil
}