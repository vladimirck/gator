package config

import (
	"encoding/json"
	"errors"
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
	config := Config{}
	currentUser, err := user.Current()

	if err != nil {
		return Config{}, err
	}

	fullPath := filepath.Join(currentUser.HomeDir, configFileName)

	//fmt.Printf("Home: %v\n", fullPath)

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
	currentUser, err := user.Current()

	if err != nil {
		return err
	}

	fullPath := filepath.Join(currentUser.HomeDir, configFileName)

	configFile, err := os.OpenFile(fullPath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return errors.New("the file " + fullPath + " cannot be opened\n")
	}
	defer configFile.Close()

	jsonConfig, err := json.Marshal(cfg)
	if err != nil {
		return errors.New("the conifuration could not be marshaled into a json format\n")
	}

	_, err = configFile.Write(jsonConfig)

	if err != nil {
		return errors.New("the configuration file could not be written into\n")
	}

	return nil
}
