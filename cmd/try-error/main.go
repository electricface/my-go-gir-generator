package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/electricface/my-go-gir-generator/config"
)

var listOnlyFlag bool

func init() {
	flag.BoolVar(&listOnlyFlag, "list-only", false, "list only")
}

func main() {
	flag.Parse()

	dir := flag.Arg(0)
	cfgFile := filepath.Join(dir, "gir-gen.toml")
	cfg, err := config.Load(cfgFile)
	if err != nil {
		log.Fatal(cfg)
	}

	for _, typeCfg := range cfg.Types {
		for _, errFunc := range typeCfg.ErrFuncs {
			log.Println(errFunc)

			if listOnlyFlag {
				continue
			}
			if handleErr(dir, typeCfg.Name, errFunc, typeCfg) {
				cfg.Save(cfgFile)
			}
		}
	}

	if !listOnlyFlag {
		test(dir, cfg)
	}
}

func getGirProjectRoot() string {
	return "github.com/linuxdeepin/go-gir"
}

func test(dir string, cfg *config.PackageConfig) {
	cmd := exec.Command("./girgen", dir, "gir-gen.toml")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	goPkg := filepath.Join(getGirProjectRoot(), strings.ToLower(cfg.Namespace)+"-"+cfg.Version)
	log.Println("go build", goPkg)
	goBuildCmd := exec.Command("go", "build", "-i", "-v", goPkg)
	goBuildCmd.Stdout = os.Stdout
	goBuildCmd.Stderr = os.Stderr

	err = goBuildCmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func handleErr(dir, typeName, funcName string, cfg *config.TypeConfig) bool {
	err := callTestOneFunc(dir, typeName, funcName)
	if err == nil {
		cfg.ErrFuncs = strSliceRemove(cfg.ErrFuncs, funcName)
		cfg.Funcs = append(cfg.Funcs, funcName)
		return true
	}
	return false
}

func strSliceRemove(s []string, a string) []string {
	result := make([]string, 0, len(s)-1)
	for _, v := range s {
		if v != a {
			result = append(result, v)
		}
	}
	return result
}

func callTestOneFunc(dir, typeName, funcName string) error {
	cmd := exec.Command("./test-one-func", dir, typeName, funcName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
