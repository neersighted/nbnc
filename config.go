package main

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"time"
)

type Config struct {
	Debug bool

	Listen string

	Cert string
	Key  string

	Auth AuthConfig

	Bouncer map[string]BouncerConfig
}

type AuthConfig struct {
	Attempts int
	Timeout  time.Duration
}

type BouncerConfig struct {
	Password string

	Bind   string
	Target string

	Secure   bool
	NoVerify bool
}

func loadConfig(file string) (config *Config, err error) {
	config = &Config{}

	// Open and read the config.
	var data []byte
	if data, err = ioutil.ReadFile(file); err != nil {
		return config, err
	}

	// Parse the config.
	if _, err = toml.Decode(string(data), config); err != nil {
		return config, err
	}

	// Set defaults.
	if config.Auth.Attempts == 0 {
		config.Auth.Attempts = 3
	}
	if config.Auth.Timeout == 0 {
		config.Auth.Timeout = 15
	}
	if config.Listen == "" {
		return config, errors.New("config: listen address must be specified")
	}
	if (config.Cert != "") != (config.Key != "") {
		return config, errors.New("config: both cert and key must be specified")
	}

	var bouncer BouncerConfig
	for i := range config.Bouncer {
		bouncer = config.Bouncer[i]
		if bouncer.Target == "" {
			return config, fmt.Errorf("config/bouncer: target address must be specified (%s)", i)
		}
	}

	return config, nil
}
