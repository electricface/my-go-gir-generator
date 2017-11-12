package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/electricface/my-go-gir-generator/config"
	"github.com/electricface/my-go-gir-generator/gi"
	"os/exec"
)

var repo *gi.Repository

func getConfigTypeMap(types []*config.TypeConfig) map[string]struct{} {
	res := make(map[string]struct{})
	for _, type0 := range types {
		res[type0.Name] = struct{}{}
	}

	return res
}

func main() {
	dir := os.Args[1]
	//targetType := os.Args[2]
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	cfgFile := filepath.Join(dir, "gir-gen.toml")
	cfg, err := config.Load(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	typeMap := getConfigTypeMap(cfg.Types)

	repo, err = gi.Load(cfg.Namespace, cfg.Version)
	if err != nil {
		log.Fatal(err)
	}

	for _, struct0 := range repo.Namespace.Structs {
		if _, ok := typeMap[struct0.Name()]; !ok {
			if shouldShowStruct(struct0) {
				callTrial(dir, struct0.Name())
			}
		}
	}

	for _, ifc := range repo.Namespace.Interfaces {
		if _, ok := typeMap[ifc.Name()]; !ok {
			callTrial(dir, ifc.Name())
		}
	}

	for _, obj := range repo.Namespace.Objects {
		if _, ok := typeMap[obj.Name()]; !ok {
			callTrial(dir, obj.Name())
		}
	}

	// repo.Namespace.Callbacks

}

func callTrial(dir, typeName string) {
	cmd := exec.Command("./trial", dir, typeName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal("trival exit with error:", err)
	}
}

func shouldShowStruct(st *gi.StructInfo) bool {
	if st.Disguised || st.GlibIsGtypeStructFor != "" {
		return false
	}
	return true
}
