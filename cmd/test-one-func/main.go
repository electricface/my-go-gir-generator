package main

import (
	"os"
	"path/filepath"

	"fmt"
	"github.com/electricface/my-go-gir-generator/config"
	"log"
	"os/exec"
	"strings"
)

const (
	normalCfgFileName = "gir-gen.toml"
	testCfgFileName   = "test-one-func.toml"
)

func main() {
	dir := os.Args[1]
	typeName := os.Args[2]
	funcName := os.Args[3]

	cfgFile := filepath.Join(dir, normalCfgFileName)

	cfg, err := config.Load(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	clearCfg(cfg, typeName, funcName)
	testCfgFile := filepath.Join(dir, testCfgFileName)
	cfg.Save(testCfgFile)

	err, exitCode := test(dir, cfg)
	if err != nil {
		log.Println(err)
		os.Exit(exitCode)
	}
}

func getGirProjectRoot() string {
	return "github.com/linuxdeepin/go-gir"
}

const (
	NoErr      = 0
	GirGenErr  = 1
	GoBuildErr = 2
)

func test(dir string, cfg *config.PackageConfig) (error, int) {
	// step 1: girgen
	girGenCmd := exec.Command("./girgen", dir, testCfgFileName)
	girGenCmd.Stdout = os.Stdout
	girGenCmd.Stderr = os.Stderr

	err := girGenCmd.Run()
	if err != nil {
		return fmt.Errorf("girgen error: %v", err), GirGenErr
	}

	// step 2: go build
	goPkg := filepath.Join(getGirProjectRoot(), strings.ToLower(cfg.Namespace)+"-"+cfg.Version)
	log.Println("go build", goPkg)
	goBuildCmd := exec.Command("go", "build", "-i", "-v", goPkg)
	goBuildCmd.Stdout = os.Stdout
	goBuildCmd.Stderr = os.Stderr

	err = goBuildCmd.Run()
	if err != nil {
		return fmt.Errorf("go build error: %v", err), GoBuildErr
	}

	// success
	return nil, NoErr
}

func clearCfg(cfg *config.PackageConfig, typeName, funcName string) {
	cfg.Funcs = nil
	var foundType bool
	for _, typ := range cfg.Types {
		typ.Funcs = nil
		typ.IgnoreFuncs = nil
		typ.ManualFuncs = nil
		typ.ErrFuncs = nil

		if typ.Name == typeName {
			foundType = true
			typ.Funcs = []string{funcName}
		}
	}

	if !foundType {
		newTypeCfg := &config.TypeConfig{
			Name:  typeName,
			Funcs: []string{funcName},
		}
		cfg.Types = append(cfg.Types, newTypeCfg)
	}
}
