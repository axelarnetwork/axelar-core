package types

type TssdConfig struct {
	Host string `mapstructure:"tssd_host"`
	Port string `mapstructure:"tssd_port"`
}
