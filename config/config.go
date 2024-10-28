package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App    `yaml:"app"`
		Hub    `yaml:"hub"`
		Room   `yaml:"room"`
		Client `yaml:"client"`
		Log    `yaml:"logger"`
	}

	App struct {
		Name string `yaml:"name"`
		Port string `yaml:"port"`
	}

	Log struct {
		Level         string `yaml:"log_level"`
		OutputPath    string `yaml:"log_output_path"`
		ErrOutputPath string `yaml:"log_err_output_path"`
	}

	Hub struct {
		BufferedRegisterSize   int `yaml:"buffered_register_size"`
		BufferedUnregisterSize int `yaml:"buffered_unregister_size"`
		BufferedMessageSize    int `yaml:"buffered_message_size"`
		BufferedRoomSize       int `yaml:"buffered_room_size"`
		BufferedEventSize      int `yaml:"buffered_event_size"`
		BufferedClientSize     int `yaml:"buffered_client_size"`
	}

	Room struct {
		BufferedRegisterSize   int `yaml:"buffered_register_size"`
		BufferedUnregisterSize int `yaml:"buffered_unregister_size"`
		BufferedMessageSize    int `yaml:"buffered_message_size"`
		BufferedClientSize     int `yaml:"buffered_client_size"`
		BufferedWorkersSize    int `yaml:"buffered_workers_size"`
	}

	Client struct {
		BufferedMessageSize int `yaml:"buffered_message_size"`
		BufferedRoomSize    int `yaml:"buffered_room_size"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	// Read config from config.yml
	cleanenv.ReadConfig("config.yaml", cfg)

	// Override config from .env file
	err := cleanenv.ReadConfig(".env", cfg)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
