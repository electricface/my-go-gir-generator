package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cosiner/gohper/terminal/std"

	"github.com/electricface/my-go-gir-generator/config"
	"github.com/electricface/my-go-gir-generator/gi"
)

var repo *gi.Repository
var girProjectRoot string

func getGirProjectRoot() string {
	return girProjectRoot
}

func setGirProjectRoot(v string) {
	girProjectRoot = v
}

func getTypeConfig(targetType string, cfg *config.PackageConfig) *config.TypeConfig {
	for _, typeCfg := range cfg.Types {
		if typeCfg.Name == targetType {
			return typeCfg
		}
	}
	return nil
}

func main() {
	dir := os.Args[1]
	targetType := os.Args[2]
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	cfgFile := filepath.Join(dir, "gir-gen.toml")
	cfg, err := config.Load(cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	goPath := os.Getenv("GOPATH")
	goSrcPrefix := filepath.Join(goPath, "src") + "/"

	if !strings.HasPrefix(dir, goSrcPrefix) {
		log.Fatalf("dir %q is not in go path %q", dir, goPath)
	}

	setGirProjectRoot(strings.TrimPrefix(filepath.Dir(dir), goSrcPrefix))
	log.Println("gir project root:", girProjectRoot)

	repo, err = gi.Load(cfg.Namespace, cfg.Version)
	if err != nil {
		log.Fatal(err)
	}

	typeDef, _ := repo.GetType(targetType)
	if typeDef == nil {
		panic("fail to get type for " + targetType)
	}

	typeCfg := getTypeConfig(targetType, cfg)
	var isNewType bool
	if typeCfg == nil {
		typeCfg = &config.TypeConfig{
			Name: targetType,
		}
		cfg.Types = append(cfg.Types, typeCfg)
		isNewType = true
	}

loop:
	for {
		funcMap := typeCfg.GetFuncMap()
		nextFunc := getNextFunc(typeDef, funcMap)
		if nextFunc == "" {
			log.Print("no next func")
			if isNewType {
				cfg.Save(cfgFile)
			}
			break
		}

		funcInfo := getFuncInfo(typeDef, nextFunc)
		if funcInfo.Deprecated {
			log.Printf("add deprecated %s to ignore_funcs", nextFunc)
			typeCfg.IgnoreFuncs = append(typeCfg.IgnoreFuncs, nextFunc)
			cfg.Save(cfgFile)
			continue
		}

		err = callTestOneFunc(dir, targetType, nextFunc)
		if err != nil {
			var interactor std.Interactor
			input := interactor.ReadInput("\nadd "+nextFunc+
				" to err_funcs(e) or manual_funcs(m) or ignore_funcs(i) or quit\n", "e")

			switch input {
			case "e":
				log.Printf("add %s to err_funcs", nextFunc)
				typeCfg.ErrFuncs = append(typeCfg.ErrFuncs, nextFunc)
				cfg.Save(cfgFile)

			case "m":
				log.Printf("add %s to manual_funcs", nextFunc)
				typeCfg.ManualFuncs = append(typeCfg.ManualFuncs, nextFunc)
				cfg.Save(cfgFile)

			case "i":
				log.Printf("add %s to ignore_funcs", nextFunc)
				typeCfg.IgnoreFuncs = append(typeCfg.IgnoreFuncs, nextFunc)
				cfg.Save(cfgFile)

			default:
				// quit
				break loop
			}
		} else {
			// err is nil
			typeCfg.Funcs = append(typeCfg.Funcs, nextFunc)
			cfg.Save(cfgFile)
		}
	}
}

func callTestOneFunc(dir, typeName, funcName string) error {
	cmd := exec.Command("./test-one-func", dir, typeName, funcName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func getNextFunc(typeDef gi.TypeDefine, funcMap map[string]int) string {
	switch td := typeDef.(type) {
	case *gi.StructInfo:
		return getStructNextFunc(td, funcMap)
	case *gi.InterfaceInfo:
		return getInterfaceNextFunc(td, funcMap)
	case *gi.ObjectInfo:
		return getObjectNextFunc(td, funcMap)
	default:
		panic(fmt.Errorf("unsupported type %T", typeDef))
	}
}

func getFuncInfo(typeDef gi.TypeDefine, funcName string) *gi.FunctionInfo {
	switch td := typeDef.(type) {
	case *gi.StructInfo:
		return getStructFuncInfo(td, funcName)
	case *gi.InterfaceInfo:
		return getInterfaceFuncInfo(td, funcName)
	case *gi.ObjectInfo:
		return getObjectFuncInfo(td, funcName)
	default:
		panic(fmt.Errorf("unsupported type %T", typeDef))
	}
}

func getStructFuncInfo(st *gi.StructInfo, funcName string) *gi.FunctionInfo {
	for _, fn := range st.Constructors {
		if fn.CIdentifier == funcName {
			return fn
		}
	}

	for _, fn := range st.Methods {
		if fn.CIdentifier == funcName {
			return fn
		}
	}

	for _, fn := range st.Functions {
		if fn.CIdentifier == funcName {
			return fn
		}
	}
	return nil
}

func getObjectFuncInfo(obj *gi.ObjectInfo, funcName string) *gi.FunctionInfo {
	for _, fn := range obj.Constructors {
		if fn.CIdentifier == funcName {
			return fn
		}
	}

	for _, fn := range obj.Methods {
		if fn.CIdentifier == funcName {
			return fn
		}
	}

	for _, fn := range obj.Functions {
		if fn.CIdentifier == funcName {
			return fn
		}
	}
	return nil
}

func getInterfaceFuncInfo(ifc *gi.InterfaceInfo, funcName string) *gi.FunctionInfo {
	for _, fn := range ifc.Methods {
		if fn.CIdentifier == funcName {
			return fn
		}
	}

	for _, fn := range ifc.Functions {
		if fn.CIdentifier == funcName {
			return fn
		}
	}
	return nil
}

func getStructNextFunc(st *gi.StructInfo, funcMap map[string]int) string {
	for _, fn := range st.Constructors {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			return fn.CIdentifier
		}
	}

	for _, fn := range st.Methods {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			return fn.CIdentifier
		}
	}

	for _, fn := range st.Functions {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			return fn.CIdentifier
		}
	}

	return ""
}

func getInterfaceNextFunc(ifc *gi.InterfaceInfo, funcMap map[string]int) string {
	for _, fn := range ifc.Methods {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			return fn.CIdentifier
		}
	}

	for _, fn := range ifc.Functions {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			return fn.CIdentifier
		}
	}

	return ""
}

func getObjectNextFunc(obj *gi.ObjectInfo, funcMap map[string]int) string {
	for _, fn := range obj.Constructors {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			return fn.CIdentifier
		}
	}

	for _, fn := range obj.Methods {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			return fn.CIdentifier
		}
	}

	for _, fn := range obj.Functions {
		if _, ok := funcMap[fn.CIdentifier]; !ok {
			return fn.CIdentifier
		}
	}

	return ""
}
