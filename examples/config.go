package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Config represents the configuration structure
type Config struct {
	KeystorePath   string `json:"keystore_path"`
	SecretKey      string `json:"secret_key"`
	AccountAddress string `json:"account_address"`
	MultiSig       struct {
		AuthorizedUsers []struct {
			Comment        string `json:"comment"`
			SecretKey      string `json:"secret_key"`
			AccountAddress string `json:"account_address"`
		} `json:"authorized_users"`
	} `json:"multi_sig"`
}

// LoadConfig loads configuration from config.json
func LoadConfig() (*Config, error) {
	configPath := "config.json"
	
	// Check if config.json exists, if not create from example
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		examplePath := "config.json.example"
		if _, err := os.Stat(examplePath); err == nil {
			fmt.Println("config.json not found. Please copy config.json.example to config.json and fill in your credentials.")
			return nil, fmt.Errorf("config.json not found")
		}
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// GetSecretKey retrieves the secret key from config or keystore
func GetSecretKey(config *Config) (string, error) {
	if config.SecretKey != "" {
		return config.SecretKey, nil
	}

	if config.KeystorePath != "" {
		keystorePath := config.KeystorePath
		if keystorePath[0] == '~' {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %v", err)
			}
			keystorePath = filepath.Join(home, keystorePath[1:])
		}

		if !filepath.IsAbs(keystorePath) {
			keystorePath = filepath.Join("examples", keystorePath)
		}

		if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
			return "", fmt.Errorf("keystore file not found: %s", keystorePath)
		}

		keystoreData, err := ioutil.ReadFile(keystorePath)
		if err != nil {
			return "", fmt.Errorf("failed to read keystore file: %v", err)
		}

		var keystore map[string]interface{}
		if err := json.Unmarshal(keystoreData, &keystore); err != nil {
			return "", fmt.Errorf("failed to parse keystore file: %v", err)
		}

		fmt.Print("Enter keystore password: ")
		// password, err := term.ReadPassword(int(syscall.Stdin))
		// if err != nil {
		//     return "", fmt.Errorf("failed to read password: %v", err)
		// }
		// fmt.Println()

		// Note: In a real implementation, you would decrypt the keystore here
		// For now, we'll return an error asking the user to use secret_key instead
		return "", fmt.Errorf("keystore decryption not implemented. Please use secret_key in config.json instead")
	}

	return "", fmt.Errorf("no secret key or keystore path provided in config")
}
