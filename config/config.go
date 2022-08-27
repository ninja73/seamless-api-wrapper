package config

import (
	"github.com/BurntSushi/toml"
	"time"
)

type Duration struct{ time.Duration }

type Server struct {
	Address      string   `toml:"address"`
	ReadTimeout  Duration `toml:"read-timeout"`
	WriteTimeout Duration `toml:"write-timeout"`
}

type Postgres struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
	PoolSize int    `toml:"pool-size"`
}

type ServerConfig struct {
	Server   Server   `toml:"server"`
	Postgres Postgres `toml:"postgres"`
}

func ParseServerConfig(configFile string) (*ServerConfig, error) {
	var config ServerConfig
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (d *Duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
