package types

// TssConfig contains all configurations values for tss
type TssConfig struct {
	Host string `mapstructure:"tofnd_host"`
	Port string `mapstructure:"tofnd_port"`
}
