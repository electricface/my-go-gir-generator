package main

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
)


type LibConfig struct {
    Namespace string `yaml:"Namespace"`
    Version string `yaml:"Version"`
    CIncludes []string `yaml:"CIncludes"`
    FunctionBlacklist []string `yaml:"FunctionBlacklist"`
    RecordBlacklist []string `yaml:"RecordBlacklist"`
    ConstantRename map[string]string `yaml:"ConstantRename"`

    functionBlacklistMap map[string]byte
    recordBlacklistMap map[string]byte
}

func (cfg *LibConfig) init() {
    cfg.functionBlacklistMap = make(map[string]byte)
    for _, f := range cfg.FunctionBlacklist {
        cfg.functionBlacklistMap[f] = 0
    }

    cfg.recordBlacklistMap = make(map[string]byte)
    for _, n := range cfg.RecordBlacklist {
        cfg.recordBlacklistMap[n] = 0
    }
}

func loadLibConfig(file string) (*LibConfig,error) {
    var libCfg LibConfig
    libCfgBytes, err := ioutil.ReadFile(file)
    if err != nil {
        return nil, err
    }

    err = yaml.Unmarshal(libCfgBytes, &libCfg)
    if err != nil {
        return nil, err
    }
    libCfg.init()
    return &libCfg, nil
}

func (cfg *LibConfig) IsFunctionInBlacklist(name string) bool {
    _, ok := cfg.functionBlacklistMap[ name ]
    return ok
}

func (cfg *LibConfig) IsRecordInBlacklist(name string) bool {
    _, ok := cfg.recordBlacklistMap[ name ]
    return ok
}
