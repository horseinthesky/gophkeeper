package server

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultEnvironment = "dev"
	defaultAddress     = "localhost:8080"
	defaultDSN         = "postgresql://postgres:mysecretpassword@localhost:15432?sslmode=disable"
	defaultClean       = 5 * time.Second
)

// Config is a gophkeeper configuration.
type Config struct {
	Environment string        `mapstructure:"ENV"`
	Address     string        `mapstructure:"ADDRESS"`
	DSN         string        `mapstructure:"DSN"`
	Clean       time.Duration `mapstructure:"CLEAN"`
}

func LoadConfig(configFilePath string) (Config, error) {
	viper.SetEnvPrefix("GOPHKEEPER")

	viper.SetDefault("ENV", defaultEnvironment)
	viper.SetDefault("ADDRESS", defaultAddress)
	viper.SetDefault("DSN", defaultDSN)
	viper.SetDefault("CLEAN", defaultClean)

	viper.SetConfigFile(configFilePath)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("%w, please provide server config file path", err)
	}

	config := Config{}
	err = viper.Unmarshal(&config)

	if config.Environment != "prod" && config.Environment != "dev" {
		return Config{}, fmt.Errorf("environment can only be dev/prod(default)")
	}

	return config, err
}
