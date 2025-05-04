package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	fmt.Printf("Hello World!\n")
	config := Config{}
	currentUser, err := user.Current()

	if err != nil {
		return Config{}, err
	}

	fullPath := filepath.Join(currentUser.HomeDir, configFileName)

	fmt.Printf("Home: %v\n", fullPath)

	configFile, err := os.Open(fullPath)
	if err != nil {
		return Config{}, err
	}
	defer configFile.Close()

	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil

}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUserName = username
	configFile, err := os.Open(configFileName)
	if err != nil {
		return err
	}
	defer configFile.Close()

	jsonConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	_, err = configFile.Write(jsonConfig)

	if err != nil {
		return err
	}

	return nil
}
