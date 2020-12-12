package types

type TssConfig struct {
	Host string `mapstructure:"tssd_host"`
	Port string `mapstructure:"tssd_port"`
}
