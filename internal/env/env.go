package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv loads the environment variables from the given file
func Load(filepath string) error {
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
