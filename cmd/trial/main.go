package main

import (
	"fmt"
	"github.com/cosiner/gohper/terminal/std"
	"github.com/electricface/my-go-gir-generator/config"
	"github.com/electricface/my-go-gir-generator/gi"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	dir := os.Args[1]
	targetType := os.Args[2]
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	cfgFile := filepath.Join(dir, "gir-gen.toml")
	cfgFileBackup := cfgFile + ".bak"
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
		funcMap := getTypeFuncMap(typeCfg)
		nextFunc := getNextFunc(typeDef, funcMap)
		if nextFunc == "" {
			log.Print("no next func")
			if isNewType {
				saveCfg(cfgFile, cfg)
			}
			break
		}

		typeCfg.Funcs = append(typeCfg.Funcs, nextFunc)

		// backup
		if err := os.Rename(cfgFile, cfgFileBackup); err != nil {
			panic(err)
		}
		saveCfg(cfgFile, cfg)

		err = test(dir, cfg)
		if err != nil {
			var interactor std.Interactor
			input := interactor.ReadInput("\nadd "+nextFunc+
				" to err_funcs(e) or manual_funcs(m), or quit\n", "e")

			switch input {
			case "m":
				log.Printf("add %s to manual_funcs", nextFunc)
				typeCfg.Funcs = typeCfg.Funcs[:len(typeCfg.Funcs)-1]
				typeCfg.ManualFuncs = append(typeCfg.ManualFuncs, nextFunc)
				saveCfg(cfgFile, cfg)

			case "e":
				log.Printf("add %s to err_funcs", nextFunc)
				typeCfg.Funcs = typeCfg.Funcs[:len(typeCfg.Funcs)-1]
				typeCfg.ErrFuncs = append(typeCfg.ErrFuncs, nextFunc)
				saveCfg(cfgFile, cfg)

			default:
				log.Println("recover config")
				if err := os.Rename(cfgFileBackup, cfgFile); err != nil {
					panic(err)
				}
				break loop

			}
		}
	}
}

func test(dir string, cfg *config.PackageConfig) error {
	output, err := exec.Command("./girgen", dir).CombinedOutput()
	os.Stdout.Write(output)
	if err != nil {
		log.Println("girgen failed:", err)
		return err
	}

	goPkg := filepath.Join(getGirProjectRoot(), strings.ToLower(cfg.Namespace)+"-"+cfg.Version)
	log.Println("go build", goPkg)
	output, err = exec.Command("go", "build", "-i", "-v", goPkg).CombinedOutput()
	os.Stdout.Write(output)
	if err != nil {
		log.Println("go build failed:", err)
		return err
	}

	return nil
}

func saveCfg(filename string, cfg *config.PackageConfig) {
	content, err := toml.Marshal(*cfg)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(filename, content, 0644)
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
