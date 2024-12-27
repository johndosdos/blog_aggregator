package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
	filename        string
}

func (c *Config) GetFilename() string {
	return c.filename
}

func (c *Config) SetUser(jsonFilenameFull, username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty: %s", username)
	}

	c.CurrentUserName = username
	file, err := openConfigFile(jsonFilenameFull)
	if err != nil {
		return err
	}
	defer file.Close()

	/*
		Encode the config struct to the JSON file. Be sure that the file has
		correct flag and permissions.
	*/
	err = json.NewEncoder(file).Encode(c)
	if err != nil {
		return fmt.Errorf("error encoding config to JSON file: %w, path: %s", err, jsonFilenameFull)
	}

	return nil
}

func openConfigFile(jsonFilenameFull string) (*os.File, error) {
	// Get home directory path.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to locate user home directory: %w.", err)
	}

	// Find .gatorconfig.json file from the home directory.
	jsonPath := homeDir + "/" + jsonFilenameFull

	// Open a file for read-write.
	// Close the file after operation (read `Read` function).
	file, err := os.OpenFile(jsonPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w, path: %s", err, jsonPath)
	}

	return file, nil
}

func Read(jsonFilenameFull string) (Config, error) {
	file, err := openConfigFile(jsonFilenameFull)
	if err != nil {
		return Config{}, err
	}
	// Close the file here.
	defer file.Close()

	// Decode the json file from the given input.
	// Set the config path
	newConfig := Config{}
	newConfig.filename = jsonFilenameFull
	err = json.NewDecoder(file).Decode(&newConfig)
	if err != nil {
		return Config{}, fmt.Errorf("unable to decode JSON file: %w", err)
	}

	return newConfig, nil
}
