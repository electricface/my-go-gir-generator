package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

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
		if shouldIgnoreFunc(funcInfo) {
			log.Printf("auto add %s to ignore_funcs", nextFunc)
			typeCfg.IgnoreFuncs = append(typeCfg.IgnoreFuncs, nextFunc)
			cfg.Save(cfgFile)
			continue
		}

		exitCode := callTestOneFunc(dir, targetType, nextFunc)
		switch exitCode {
		case 0:
			// no err
			typeCfg.Funcs = append(typeCfg.Funcs, nextFunc)

		case 1:
			typeCfg.ErrFuncs = append(typeCfg.ErrFuncs, nextFunc)
			log.Printf("auto add %s to err_funcs", nextFunc)

		default:
			// 2
			notifyUser("require manual intervention")
			var interactor std.Interactor
			input := interactor.ReadInput("\nadd "+nextFunc+
				" to err_funcs(e) or manual_funcs(m) or ignore_funcs(i) or quit\n", "e")

			switch input {
			case "e":
				log.Printf("add %s to err_funcs", nextFunc)
				typeCfg.ErrFuncs = append(typeCfg.ErrFuncs, nextFunc)

			case "m":
				log.Printf("add %s to manual_funcs", nextFunc)
				typeCfg.ManualFuncs = append(typeCfg.ManualFuncs, nextFunc)

			case "i":
				log.Printf("add %s to ignore_funcs", nextFunc)
				typeCfg.IgnoreFuncs = append(typeCfg.IgnoreFuncs, nextFunc)

			default:
				// quit
				break loop
			}
		}
		cfg.Save(cfgFile)
	}
}

func shouldIgnoreFunc(fn *gi.FunctionInfo) bool {
	if fn.Deprecated {
		return true
	}

	if fn.Parameters != nil {
		for _, param := range fn.Parameters.Parameters {
			if param.Name == "..." {
				return true
			}

			if param.Type != nil && param.Type.Name == "va_list" {
				return true
			}
		}
	}

	return false
}

func notifyUser(msg string) {
	exec.Command("notify-send", "trial", msg).Run()
}

func callTestOneFunc(dir, typeName, funcName string) int {
	cmd := exec.Command("./test-one-func", dir, typeName, funcName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err == nil {
		return 0
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		log.Fatal("err is not ExitError")
	}
	waitStatus := exitErr.ProcessState.Sys().(syscall.WaitStatus)
	return waitStatus.ExitStatus()
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
