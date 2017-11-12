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
	Callbacks []string      `toml:"callbacks"`
}

type TypeConfig struct {
	Name        string   `toml:"name"`
	Funcs       []string `toml:"funcs"`
	ManualFuncs []string `toml:"manual_funcs"`
	ErrFuncs    []string `toml:"err_funcs"`
	IgnoreFuncs []string `toml:"ignore_funcs"`
}

const (
	NormalFunc = iota
	ErrFunc
	ManualFunc
	IgnoreFunc
)

func (typeCfg *TypeConfig) GetFuncMap() map[string]int {
	ret := make(map[string]int, len(typeCfg.Funcs)+len(typeCfg.ErrFuncs)+
		len(typeCfg.ManualFuncs)+len(typeCfg.IgnoreFuncs))
	for _, fn := range typeCfg.Funcs {
		if _, ok := ret[fn]; ok {
			panic("duplicated func " + fn)
		}
		ret[fn] = NormalFunc
	}
	for _, fn := range typeCfg.ErrFuncs {
		if _, ok := ret[fn]; ok {
			panic("duplicated func " + fn)
		}
		ret[fn] = ErrFunc
	}
	for _, fn := range typeCfg.ManualFuncs {
		if _, ok := ret[fn]; ok {
			panic("duplicated func " + fn)
		}
		ret[fn] = ManualFunc
	}
	for _, fn := range typeCfg.IgnoreFuncs {
		if _, ok := ret[fn]; ok {
			panic("duplicated func " + fn)
		}
		ret[fn] = IgnoreFunc
	}
	return ret
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
