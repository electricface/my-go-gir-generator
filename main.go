package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"mygi"
)

var repo *mygi.Repository
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

	cfg, err := LoadConfig(filepath.Join(dir, "gir-gen.toml"))
	if err != nil {
		log.Fatal(err)
	}
	goPath := os.Getenv("GOPATH")
	goSrcPrefix := filepath.Join(goPath, "src") + "/"

	if !strings.HasPrefix(dir, goSrcPrefix) {
		log.Fatalf("dir %q is not in go path %q", dir, goPath)
	}

	setGirProjectRoot(strings.TrimPrefix(filepath.Dir(dir), goSrcPrefix))

	repo, err = mygi.Load(cfg.Namespace, cfg.Version)
	if err != nil {
		log.Fatal(err)
	}

	types := repo.GetTypes()
	log.Print(len(types))

	for _, genFileCfg := range cfg.GenFiles {
		typeDef, ns := repo.GetType(genFileCfg.Type)
		if typeDef == nil {
			panic("fail to get type for " + genFileCfg.Type)
		}
		if ns != cfg.Namespace {
			panic("assert failed ns == cfg.Namespace")
		}

		pkg := strings.ToLower(cfg.Namespace)
		sourceFile := NewSourceFile(pkg)

		switch td := typeDef.(type) {
		case *mygi.StructInfo:
			pStruct(sourceFile, td, genFileCfg.Funcs)
		case *mygi.InterfaceInfo:
			pInterface(sourceFile, td, genFileCfg.Funcs)
		case *mygi.ObjectInfo:
			pObject(sourceFile, td, genFileCfg.Funcs)
		}

		// cgo pkg-config
		for _, pkg := range repo.Packages {
			sourceFile.AddCPkg(pkg.Name)
		}

		// c header files
		for _, cInc := range repo.CIncludes() {
			sourceFile.AddCInclude("<" + cInc.Name + ">")
		}

		outFile := filepath.Join(dir, getOutputFileBaseName(genFileCfg))
		log.Println("outFile:", outFile)
		sourceFile.Save(outFile)
	}

	//repo.GetType()
	//
	//for name, type0 := range types {
	//	log.Printf("%s -> %T\n", name, type0)
	//}
}

func getOutputFileBaseName(cfg *GenFileConfig) string {
	var name string
	if cfg.Filename != "" {
		name = cfg.Filename
	} else {
		name = camel2Snake(cfg.Type)
	}
	return name + "_auto.go"
}

func pStruct(s *SourceFile, struct0 *mygi.StructInfo, funcs []string) {
	s.AddGoImport("unsafe")
	name := struct0.Name()
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

func pObject(s *SourceFile, class *mygi.ObjectInfo, funcs []string) {
	s.AddGoImport("unsafe")
	name := class.Name()
	s.GoBody.Pn("// Object %s", name)

	s.GoBody.Pn("type %s struct {", name)
	s.GoBody.Pn("Ptr unsafe.Pointer")
	s.GoBody.Pn("}")

	cPtrType := "*C." + class.CTypeAttr

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
	for _, method := range class.Methods {
		if strSliceContains(funcs, method.CIdentifier) {
			pFunction(s, method)
		}
	}
}

func pInterface(s *SourceFile, ifc *mygi.InterfaceInfo, funcs []string) {
	s.AddGoImport("unsafe")
	name := ifc.Name()
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

func pFunction(s *SourceFile, method *mygi.FunctionInfo) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("pFunction", method.CIdentifier)
			panic(err)
		}
	}()
	//spew.Dump(method)
	s.GoBody.Pn("// %s is a wrapper around %s().", method.Name(), method.CIdentifier)

	var receiver string
	var args []string
	var instanceParamTpl *ParamPassInTemplate
	var paramTpls []*ParamPassInTemplate

	if method.Parameters != nil {
		instanceParam := method.Parameters.InstanceParameter
		if instanceParam != nil {
			instanceParamTpl = newParamPassInTemplate(instanceParam)
			receiver = "(" + instanceParamTpl.GetVarTypeForGo() + ")"
		}

		for _, param := range method.Parameters.Parameters {
			tpl := newParamPassInTemplate(param)
			paramTpls = append(paramTpls, tpl)
			args = append(args, tpl.GetVarTypeForGo())
		}
	}

	argsJoined := strings.Join(args, ", ")

	var retTypes []string
	var retValTpl *ReturnValueTemplate
	if method.ReturnValue.Type.Name != "none" {
		retValTpl = newReturnValueTemplate(method.ReturnValue)
		retTypes = append(retTypes, retValTpl.GetTypeForGo())
	}
	if method.Throws {
		retTypes = append(retTypes, "error")
	}

	retTypesJoined := strings.Join(retTypes, ", ")
	if strings.Contains(retTypesJoined, ",") {
		retTypesJoined = "(" + retTypesJoined + ")"
	}
	s.GoBody.Pn("func %s %s (%s) %s {", receiver, method.Name(), argsJoined, retTypesJoined)

	// start func body
	var exprsInCall []string
	if instanceParamTpl != nil {
		instanceParamTpl.WriteDeclaration(s)
		exprsInCall = append(exprsInCall, instanceParamTpl.GetExprInCall())
	}

	for _, paramTpl := range paramTpls {
		paramTpl.WriteDeclaration(s)
	}

	if method.Throws {
		s.AddGirImport("glib")
		s.GoBody.Pn("var err glib.Error")
	}

	for _, paramTpl := range paramTpls {
		exprsInCall = append(exprsInCall, paramTpl.GetExprInCall())
	}
	if method.Throws {
		exprsInCall = append(exprsInCall, "(**C.GError)(unsafe.Pointer(&err))")
	}

	call := fmt.Sprintf("C.%s(%s)", method.CIdentifier, strings.Join(exprsInCall, ", "))
	if retValTpl != nil {
		s.GoBody.P("ret0 := ")
	}
	s.GoBody.Pn(call)

	if retValTpl != nil {
		if method.Throws {
			s.GoBody.Pn("if err.Ptr != nil {")
			s.GoBody.Pn("defer err.Free()")
			s.GoBody.Pn("return %s, err.GoValue()", retValTpl.ErrorReturn())
			s.GoBody.Pn("}")
			retValTpl.WriteClean(s)
			s.GoBody.Pn("return %s,nil", retValTpl.NormalReturn())
		} else {
			retValTpl.WriteClean(s)
			s.GoBody.Pn("return %s", retValTpl.NormalReturn())
		}

	} else {
		// retValTpl is nil
		if method.Throws {
			s.GoBody.Pn("if err.Ptr != nil {")
			s.GoBody.Pn("defer err.Free()")
			s.GoBody.Pn("return err.GoValue()")
			s.GoBody.Pn("}")
			s.GoBody.Pn("return nil")
		}
	}

	s.GoBody.Pn("}") // end body
}
