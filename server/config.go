package server

import "github.com/spf13/viper"

const (
	defaultAddress = "localhost:8080"
	defaultDSN     = "postgresql://postgres@localhost:5432?sslmode=disable"
)

// Config is a gophkeeper configuration.
type Config struct {
	Address string `mapstructure:"ADDRESS"`
	DSN     string `mapstructure:"DSN"`
}

func LoadConfig(path string) (Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = viper.Unmarshal(&config)

	if config.Address == "" {
		config.Address = defaultAddress
	}

	if config.DSN == "" {
		config.DSN = defaultDSN
	}

	return config, err
}
