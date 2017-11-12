package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/electricface/my-go-gir-generator/config"
	"github.com/electricface/my-go-gir-generator/gi"
)

var repo *gi.Repository
var listOnlyFlag bool

func init() {
	flag.BoolVar(&listOnlyFlag, "list-only", false, "list only")
}

func getConfigTypeMap(types []*config.TypeConfig) map[string]*config.TypeConfig {
	res := make(map[string]*config.TypeConfig)
	for _, type0 := range types {
		res[type0.Name] = type0
	}
	return res
}

const (
	NormalFunc = iota
	ErrFunc
	ManualFunc
	IgnoreFunc
)

func getTypeFuncMap(typeCfg *config.TypeConfig) map[string]int {
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

func main() {
	flag.Parse()

	dir := flag.Arg(0)
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
		if typeCfg, ok := typeMap[struct0.Name()]; !ok {
			if shouldShowStruct(struct0) {
				callTrial(dir, struct0.Name())
			}
		} else {
			funcMap := getTypeFuncMap(typeCfg)
			need := listMissingFuncsStruct(struct0, funcMap)
			if need {
				callTrial(dir, struct0.Name())
			}
		}
	}

	for _, ifc := range repo.Namespace.Interfaces {
		if typeCfg, ok := typeMap[ifc.Name()]; !ok {
			callTrial(dir, ifc.Name())
		} else {
			funcMap := getTypeFuncMap(typeCfg)
			need := listMissingFuncsInterface(ifc, funcMap)
			if need {
				callTrial(dir, ifc.Name())
			}
		}
	}

	for _, obj := range repo.Namespace.Objects {
		if typeCfg, ok := typeMap[obj.Name()]; !ok {
			callTrial(dir, obj.Name())
		} else {
			funcMap := getTypeFuncMap(typeCfg)
			need := listMissingFuncsObject(obj, funcMap)
			if need {
				callTrial(dir, obj.Name())
			}
		}
	}

	// repo.Namespace.Callbacks

}

func listMissingFuncsStruct(st *gi.StructInfo, funcMap map[string]int) (need bool) {
	for _, fn := range st.Constructors {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			need = true
			fmt.Println(fn.CIdentifier)
		}
	}

	for _, fn := range st.Methods {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			need = true
			fmt.Println(fn.CIdentifier)
		}
	}

	for _, fn := range st.Functions {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			need = true
			fmt.Println(fn.CIdentifier)
		}
	}
	return
}

func listMissingFuncsObject(obj *gi.ObjectInfo, funcMap map[string]int) (need bool) {
	for _, fn := range obj.Constructors {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			need = true
			fmt.Println(fn.CIdentifier)
		}
	}

	for _, fn := range obj.Methods {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			need = true
			fmt.Println(fn.CIdentifier)
		}
	}

	for _, fn := range obj.Functions {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			need = true
			fmt.Println(fn.CIdentifier)
		}
	}

	return
}

func listMissingFuncsInterface(ifc *gi.InterfaceInfo, funcMap map[string]int) (need bool) {
	for _, fn := range ifc.Methods {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			need = true
			fmt.Println(fn.CIdentifier)
		}
	}

	for _, fn := range ifc.Functions {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			need = true
			fmt.Println(fn.CIdentifier)
		}
	}

	return
}

func callTrial(dir, typeName string) {
	if listOnlyFlag {
		fmt.Println(typeName)
		return
	}

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
