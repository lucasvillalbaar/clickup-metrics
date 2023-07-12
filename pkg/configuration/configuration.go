package configuration

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type EnvVars struct {
	Token  string
	ApiKey string
}

var envVars EnvVars

// getEnvVariable retrieves the value of the specified environment variable.
// If the variable is not set, it logs an error and terminates the program.
func getEnvVariable(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Error: Environment variable %s is not set", key)
	}
	return value
}

// GetEnvironmentVariables returns the loaded environment variables.
func GetEnvironmentVariables() EnvVars {
	return envVars
}

// LoadEnvironmentVariables loads the environment variables from the .env file.
// It returns an error if there was a problem loading the file.
func LoadEnvironmentVariables() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}
	envVars = EnvVars{
		Token:  getEnvVariable("TOKEN"),
		ApiKey: getEnvVariable("API_KEY"),
	}

	return nil
}
