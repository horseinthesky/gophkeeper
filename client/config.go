package client

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultEnvironment = "prod"
	defaultAddress     = "localhost:8080"
	defaultDSN         = "postgresql://postgres:mysecretpassword@localhost:25432?sslmode=disable"
	defaultSync        = 15 * time.Second
	defaultClean       = 5 * time.Second
)

// Config is a gophkeeper configuration.
type Config struct {
	User        string        `mapstructure:"USER"`
	Password    string        `mapstructure:"PASSWORD"`
	Environment string        `mapstructure:"ENV"`
	Address     string        `mapstructure:"ADDRESS"`
	DSN         string        `mapstructure:"DSN"`
	Sync        time.Duration `mapstructure:"SYNC"`
	Clean       time.Duration `mapstructure:"CLEAN"`
}

func LoadConfig(path string) (Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOPHKEEPER")

	viper.SetDefault("ENV", defaultEnvironment)
	viper.SetDefault("ADDRESS", defaultAddress)
	viper.SetDefault("DSN", defaultDSN)
	viper.SetDefault("SYNC", defaultSync)
	viper.SetDefault("CLEAN", defaultClean)
	viper.SetDefault("PASSWORD", "")

	if path != "" {
		viper.SetConfigFile(path)

		err := viper.ReadInConfig()
		if err != nil {
			return Config{}, fmt.Errorf("%w, please provide server config file path", err)
		}
	}

	config := Config{}
	err := viper.Unmarshal(&config)

	if config.Environment != "prod" && config.Environment != "dev" {
		return Config{}, fmt.Errorf("environment can only be dev/prod(default)")
	}

	if config.User == "" {
		return Config{}, fmt.Errorf("user name cannot be empty")
	}

	if config.Password == "" {
		return Config{}, fmt.Errorf("user password cannot be empty")
	}

	return config, err
}
