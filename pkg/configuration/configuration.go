/*
Package configuration provides functionality for loading and accessing environment variables.

This package is responsible for loading environment variables from a .env file and making them accessible
to other parts of the application. It provides a way to retrieve the loaded environment variables and
terminate the program if any required variables are missing.

Usage:
 1. Call LoadEnvironmentVariables to load the environment variables from the .env file.
 2. Use GetEnvironmentVariables to retrieve the loaded variables.
 3. Access individual variables as needed.

Example:

	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	envVars := configuration.GetEnvironmentVariables()
	fmt.Println("Token:", envVars.Token)
	fmt.Println("API Key:", envVars.ApiKey)
*/
package configuration

import (
	"fmt"
	"log"
	"os"
)

type EnvVars struct {
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
	envVars = EnvVars{
		ApiKey: getEnvVariable("API_KEY"),
	}
	if envVars.ApiKey == "" {
		return fmt.Errorf("API_KEY not found")
	}

	return nil
}
