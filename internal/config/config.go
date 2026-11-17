// Package config loads environment configuration (the .env file).
package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// LoadDotEnv reads the .env file and sets environment variables,
// if they are not already set from the environment
func LoadDotEnv() error {
	err := godotenv.Load()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
