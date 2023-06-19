package config

import (
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/configloader"
	"gopkg.in/yaml.v3"
)

const configFilepath = "./config.yaml"

type HealthCheckConfig struct {
	Port               int    `yaml:"healthcheck_port"`
	PacketLimit        int    `yaml:"packet_limit"`
	Protocol           string `yaml:"protocol"`
	HealthCheckMessage string `yaml:"healthcheck_message"`
	HealthCheckACK     string `yaml:"healthcheck_response"`
}

func LoadConfig() (*HealthCheckConfig, error) {
	configFile, err := configloader.GetConfigFileAsBytes(configFilepath)
	if err != nil {
		return nil, err
	}

	var healthCheckerConfig HealthCheckConfig
	err = yaml.Unmarshal(configFile, &healthCheckerConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing health checker config file: %s", err)
	}

	return &healthCheckerConfig, nil
}
