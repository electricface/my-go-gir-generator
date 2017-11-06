package config

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
)

type PackageConfig struct {
	Namespace string        `toml:"namespace"`
	Version   string        `toml:"version"`
	Types     []*TypeConfig `toml:"types"`
	Funcs     []string      `toml:"funcs"`
}

type TypeConfig struct {
	Name        string   `toml:"name"`
	Funcs       []string `toml:"funcs"`
	ManualFuncs []string `toml:"manual_funcs"`
	ErrFuncs    []string `toml:"err_funcs"`
	IgnoreFuncs []string `toml:"ignore_funcs"`
}

func Load(filename string) (*PackageConfig, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config PackageConfig
	err = toml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
