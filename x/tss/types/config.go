package types

import "time"

// TssConfig contains all configurations values for tss
type TssConfig struct {
	Host        string        `mapstructure:"tofnd-host"`
	Port        string        `mapstructure:"tofnd-port"`
	DialTimeout time.Duration `mapstructure:"tofnd-dial-timeout"`
}

// DefaultConfig returns the default tss configuration
func DefaultConfig() TssConfig {
	return TssConfig{
		Host:        "localhost",
		Port:        "50051",
		DialTimeout: 15 * time.Second,
	}
}
