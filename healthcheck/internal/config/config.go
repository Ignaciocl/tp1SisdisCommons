package config

type HealthCheckConfig struct {
	Port               int
	PacketLimit        int
	Protocol           string
	HealthCheckMessage string
	HealthCheckACK     string
}

func GetHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		Port:               6969,
		PacketLimit:        8 * 1024,
		Protocol:           "tcp",
		HealthCheckMessage: "te estas portando mal",
		HealthCheckACK:     "seras castigada",
	}
}
