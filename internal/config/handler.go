package config

type Handler struct {
	DefaultOffset int `mapstructure:"default_offset"`
	DefaultLimit  int `mapstructure:"default_limit"`
}
