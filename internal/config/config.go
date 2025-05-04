package config

import (
	"encoding/json"
	"os"
)

const configFileName = "~/.gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	config := Config{}

	configFile, err := os.Open(configFileName)
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
