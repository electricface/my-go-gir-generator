package main

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
)

type Config struct {
	Namespace string           `toml:"namespace"`
	Version   string           `toml:"version"`
	GenFiles  []*GenFileConfig `toml:"gen_files"`
	Funcs     []string         `toml:"funcs"`
}

type GenFileConfig struct {
	Type     string   `toml:"type"`
	Filename string   `toml:"filename"`
	Funcs    []string `toml:"funcs"`
}

func LoadConfig(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = toml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
