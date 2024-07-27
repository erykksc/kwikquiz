package config

import (
	"bufio"
	"os"
	"strings"
)

// GlobalConfig is the global configuration,
// it should be set by the main function
type Config struct {
	config
}

type config struct {
	InDevMode  bool
	InProdMode bool
	// DBUser     string
	// DBPassword string
	// DBName     string
	// DBHost     string
	// DBPort     string
}

// LoadConfigFromEnv loads the configuration from the environment variables
func LoadConfigFromEnv() (Config, error) {
	prodVar := os.Getenv("PROD")
	inProd := prodVar == "true" || prodVar == "1"
	inDebug := !inProd
	cfg := Config{
		config{
			InDevMode:  inDebug,
			InProdMode: inProd,
			// DBUser:     os.Getenv("DB_USER"),
			// DBPassword: os.Getenv("DB_PASSWORD"),
			// DBName:     os.Getenv("DB_NAME"),
			// DBHost:     os.Getenv("DB_HOST"),
			// DBPort:     os.Getenv("DB_PORT"),
		},
	}

	// Check if the environment variables are set
	// if cfg.DBUser == "" {
	// 	return cfg, fmt.Errorf("DB_USER is not set")
	// }
	// if cfg.DBPassword == "" {
	// 	return cfg, fmt.Errorf("DB_PASSWORD is not set")
	// }
	// if cfg.DBName == "" {
	// 	return cfg, fmt.Errorf("DB_NAME is not set")
	// }
	// if cfg.DBHost == "" {
	// 	return cfg, fmt.Errorf("DB_HOST is not set")
	// }
	// if cfg.DBPort == "" {
	// 	return cfg, fmt.Errorf("DB_PORT is not set")
	// }

	return cfg, nil
}

// LoadEnv loads the environment variables from the given file
func LoadEnv(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.SplitN(line, "=", 2)
		if len(split) == 2 {
			key := split[0]
			value := split[1]
			os.Setenv(key, value)
		}
	}
	return nil
}
