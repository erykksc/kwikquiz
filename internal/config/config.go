package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	DBUser     string `ini:"DBUser"`
	DBPassword string `ini:"database.DBPassword"`
	DBName     string `ini:"database.DBName"`
	DBHost     string `ini:"database.DBHost"`
	DBPort     string `ini:"database.DBPort"`
}

func LoadConfig() (*Config, error) {

	viper.SetConfigName("de_local")
	viper.SetConfigType("properties")
	viper.AddConfigPath("./")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil

}
