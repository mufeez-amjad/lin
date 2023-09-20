package config

import (
	"log"
	"os"
	"sync"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

type Config struct {
	APIKey          string `mapstructure:"apikey"`
	GraphQLEndpoint string `mapstructure:"graphqlendpoint"`
}

var (
	config     Config
	initOnce   sync.Once
	configDir  = xdg.ConfigHome + "/lin/"
	configName = "config"
	configPath = configDir + configName + ".yaml"
)

func GetConfig() Config {
	initOnce.Do(func() {
		viper.SetConfigName(configName)
		viper.SetConfigType("yaml")
		viper.AddConfigPath(configDir)
		viper.AddConfigPath(".")

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error if desired
			} else {
				log.Fatalf("Failed to read config file %v", err)
			}
		}
		// Defaults
		viper.SetDefault("GraphQLEndpoint", "https://api.linear.app/graphql")

		// Unmarshal the configuration into the struct
		if err := viper.Unmarshal(&config); err != nil {
			panic(err)
		}
	})

	return config
}

func SaveConfig() {
	config.Save()
}

func (c Config) Save() {
	if _, err := os.Stat(configDir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(configDir, 0700)
		}
	}

	defer func() {
		if err := viper.WriteConfigAs(configPath); err != nil {
			if os.IsNotExist(err) {
				err = viper.WriteConfigAs(configPath)
				if err != nil {
					log.Fatalf("%v", err)
				}
			} else {
				log.Fatalf("%v", err)
			}
		}
	}()

	viper.Set("APIKey", config.APIKey)
	viper.Set("GraphQLEndpoint", config.GraphQLEndpoint)
}
