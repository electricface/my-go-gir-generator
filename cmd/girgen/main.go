package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

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

func main() {
	dir := os.Args[1]
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.Load(filepath.Join(dir, "gir-gen.toml"))
	if err != nil {
		log.Fatal(err)
	}
	goPath := os.Getenv("GOPATH")
	goSrcPrefix := filepath.Join(goPath, "src") + "/"

	if !strings.HasPrefix(dir, goSrcPrefix) {
		log.Fatalf("dir %q is not in go path %q", dir, goPath)
	}

	setGirProjectRoot(strings.TrimPrefix(filepath.Dir(dir), goSrcPrefix))

	repo, err = gi.Load(cfg.Namespace, cfg.Version)
	if err != nil {
		log.Fatal(err)
	}

	pkg := strings.ToLower(repo.Namespace.Name)
	sourceFile := getSourceFile(repo, pkg)

	for _, genTypeCfg := range cfg.Types {
		typeDef, ns := repo.GetType(genTypeCfg.Name)
		if typeDef == nil {
			panic("fail to get type for " + genTypeCfg.Name)
		}
		if ns != cfg.Namespace {
			panic("assert failed ns == cfg.Namespace")
		}

		switch td := typeDef.(type) {
		case *gi.StructInfo:
			pStruct(sourceFile, td, genTypeCfg.Funcs)
		case *gi.InterfaceInfo:
			pInterface(sourceFile, td, genTypeCfg.Funcs)
		case *gi.ObjectInfo:
			pObject(sourceFile, td, genTypeCfg.Funcs)
		}
	}

	for _, fn := range repo.Namespace.Functions {
		if strSliceContains(cfg.Funcs, fn.CIdentifier) {
			pFunction(sourceFile, fn)
		}
	}

	// alias
	for _, alias := range repo.Namespace.Aliases {
		pAlias(sourceFile, alias)
	}

	// enums
	for _, enum := range repo.Namespace.Enums {
		pEnum(sourceFile, enum)
	}
	for _, enum := range repo.Namespace.Bitfields {
		pEnum(sourceFile, enum)
	}

	outFile := filepath.Join(dir, pkg+"_auto.go")
	log.Println("outFile:", outFile)
	sourceFile.Save(outFile)
}

func getSourceFile(repo *gi.Repository, pkg string) *SourceFile {
	sourceFile := NewSourceFile(pkg)

	// cgo pkg-config
	for _, pkg := range repo.Packages {
		sourceFile.AddCPkg(pkg.Name)
	}

	// c header files
	for _, cInc := range repo.CIncludes() {
		sourceFile.AddCInclude("<" + cInc.Name + ">")
	}
	return sourceFile
}

func pAlias(s *SourceFile, alias *gi.AliasInfo) {
	cType := alias.CType()
	s.GoBody.Pn("type %s %s", alias.Name(), cType.CgoNotation())
}

func pEnum(s *SourceFile, enum *gi.EnumInfo) {
	name := enum.Name()
	s.GoBody.Pn("type %s int", name)
	s.GoBody.Pn("const (")
	for i, member := range enum.Members {
		memberName := name + snake2Camel(member.Name)
		if i == 0 {
			s.GoBody.Pn("%s %s = %s", memberName, name, member.Value)
		} else {
			s.GoBody.Pn("%s = %s", memberName, member.Value)
		}
	}
	s.GoBody.Pn(")") // end const
}

func pStruct(s *SourceFile, struct0 *gi.StructInfo, funcs []string) {
	s.AddGoImport("unsafe")
	name := struct0.Name()
	defer func() {
		if err := recover(); err != nil {
			log.Println("pStruct", name)
			panic(err)
		}
	}()
	s.GoBody.Pn("// Struct %s", name)

	s.GoBody.Pn("type %s struct {", name)
	s.GoBody.Pn("Ptr unsafe.Pointer")
	s.GoBody.Pn("}")

	cPtrType := "*C." + struct0.CTypeAttr

	// method native
	s.GoBody.Pn("func (v %s) native() %s {", name, cPtrType)
	s.GoBody.Pn("return (%s)(v.Ptr)", cPtrType)
	s.GoBody.Pn("}")

	// method wrapXXX
	s.GoBody.Pn("func wrap%s(p %s) %s {", name, cPtrType, name)
	s.GoBody.Pn("return %s{unsafe.Pointer(p)}", name)
	s.GoBody.Pn("}")

	// method WrapXXX
	s.GoBody.Pn("func Wrap%s(p unsafe.Pointer) %s {", name, name)
	s.GoBody.Pn("return %s{p}", name)
	s.GoBody.Pn("}")

	// constructors
	for _, fn := range struct0.Constructors {
		if strSliceContains(funcs, fn.CIdentifier) {
			pFunction(s, fn)
		}
	}

	// methods
	for _, method := range struct0.Methods {
		if strSliceContains(funcs, method.CIdentifier) {
			pFunction(s, method)
		}
	}

	// functions
	for _, fn := range struct0.Functions {
		if strSliceContains(funcs, fn.CIdentifier) {
			pFunction(s, fn)
		}
	}
}

func pObject(s *SourceFile, object *gi.ObjectInfo, funcs []string) {
	s.AddGoImport("unsafe")
	name := object.Name()
	defer func() {
		if err := recover(); err != nil {
			log.Println("pObject", name)
			panic(err)
		}
	}()
	s.GoBody.Pn("// Object %s", name)

	if object.Parent != "" {
		parent, parentNS := repo.GetType(object.Parent)
		if parent == nil {
			panic("fail to get type " + object.Parent)
		}

		parentNSLower := strings.ToLower(parentNS)
		s.GoBody.Pn("type %s struct {", name)
		if isSameNamespace(parentNS) {
			s.GoBody.Pn("%s", parent.Name())
		} else {
			s.AddGirImport(parentNS)
			s.GoBody.Pn("%s.%s", parentNSLower, parent.Name())
		}
		s.GoBody.Pn("}")
	} else {
		s.GoBody.Pn("type %s struct {", name)
		s.GoBody.Pn("Ptr unsafe.Pointer")
		s.GoBody.Pn("}")
	}

	cPtrType := "*C." + object.CTypeAttr

	// method native
	s.GoBody.Pn("func (v %s) native() %s {", name, cPtrType)
	s.GoBody.Pn("return (%s)(v.Ptr)", cPtrType)
	s.GoBody.Pn("}")

	// method wrapXXX
	s.GoBody.Pn("func wrap%s(p %s) (v %s) {", name, cPtrType, name)
	s.GoBody.Pn("v.Ptr = unsafe.Pointer(p)")
	s.GoBody.Pn("return")
	s.GoBody.Pn("}")

	// method WrapXXX
	s.GoBody.Pn("func Wrap%s(p unsafe.Pointer) (v %s) {", name, name)
	s.GoBody.Pn("v.Ptr = p")
	s.GoBody.Pn("return")
	s.GoBody.Pn("}")

	for _, ifc0 := range object.ImplementedInterfaces() {
		ifc, ifcNS := repo.GetType(ifc0)
		if ifc == nil {
			panic("fail to get type " + ifc0)
		}
		ifcInfo := ifc.(*gi.InterfaceInfo)

		ifcNSLower := strings.ToLower(ifcNS)
		// method name is ifcInfo.Name()
		if isSameNamespace(ifcNS) {
			s.GoBody.Pn("func (v %s) %s() %s {", name, ifcInfo.Name(), ifcInfo.Name())
			s.GoBody.Pn("    return Wrap%s(v.Ptr)", ifcInfo.Name())
		} else {
			s.GoBody.Pn("func (v %s) %s() %s.%s {", name, ifcInfo.Name(), ifcNSLower, ifcInfo.Name())
			s.GoBody.Pn("    return %s.Wrap%s(v.Ptr) /*gir:%s*/", ifcNSLower, ifcInfo.Name(), ifcNS)
		}
		s.GoBody.Pn("}")
	}

	// constructors
	for _, fn := range object.Constructors {
		if strSliceContains(funcs, fn.CIdentifier) {
			pFunction(s, fn)
		}
	}

	// methods
	for _, method := range object.Methods {
		if strSliceContains(funcs, method.CIdentifier) {
			pFunction(s, method)
		}
	}

	// functions
	for _, fn := range object.Functions {
		if strSliceContains(funcs, fn.CIdentifier) {
			pFunction(s, fn)
		}
	}
}

func pInterface(s *SourceFile, ifc *gi.InterfaceInfo, funcs []string) {
	s.AddGoImport("unsafe")
	name := ifc.Name()
	defer func() {
		if err := recover(); err != nil {
			log.Println("pInterface", name)
			panic(err)
		}
	}()
	s.GoBody.Pn("// Interface %s", name)

	s.GoBody.Pn("type %s struct {", name)
	s.GoBody.Pn("Ptr unsafe.Pointer")
	s.GoBody.Pn("}")

	cPtrType := "*C." + ifc.CTypeAttr

	// method native
	s.GoBody.Pn("func (v %s) native() %s {", name, cPtrType)
	s.GoBody.Pn("return (%s)(v.Ptr)", cPtrType)
	s.GoBody.Pn("}")

	// method wrapXXX
	s.GoBody.Pn("func wrap%s(p %s) %s {", name, cPtrType, name)
	s.GoBody.Pn("return %s{unsafe.Pointer(p)}", name)
	s.GoBody.Pn("}")

	// method WrapXXX
	s.GoBody.Pn("func Wrap%s(p unsafe.Pointer) %s {", name, name)
	s.GoBody.Pn("return %s{p}", name)
	s.GoBody.Pn("}")

	// methods
	for _, method := range ifc.Methods {
		if strSliceContains(funcs, method.CIdentifier) {
			pFunction(s, method)
		}
	}
}
