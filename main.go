package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"mygi"
)

var repo *mygi.Repository

func main() {
	dir := os.Args[1]
	cfg, err := LoadConfig(filepath.Join(dir, "gir-gen.toml"))
	if err != nil {
		log.Fatal(err)
	}

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
		case *mygi.Interface:
			pInterface(sourceFile, td, genFileCfg.Funcs)
		case *mygi.Class:
			pClass(sourceFile, td, genFileCfg.Funcs)
		}

		outFile := filepath.Join(dir, genFileCfg.Filename+"_auto.go")
		log.Println("outFile:", outFile)
		sourceFile.Save(outFile)
	}

	//repo.GetType()
	//
	//for name, type0 := range types {
	//	log.Printf("%s -> %T\n", name, type0)
	//}
}

func pClass(s *SourceFile, class *mygi.Class, funcs []string) {
	name := class.Name()
	s.GoBody.Pn("// class %s", name)

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
	// methods
	for _, method := range class.Methods {
		if strSliceContains(funcs, method.CIdentifier) {
			pMethod(s, method)
		}
	}
}

func pInterface(s *SourceFile, interface0 *mygi.Interface, funcs []string) {
	name := interface0.Name()
	s.GoBody.Pn("// interface %s", name)

	s.GoBody.Pn("type %s struct {", name)
	s.GoBody.Pn("Ptr unsafe.Pointer")
	s.GoBody.Pn("}")

	cPtrType := "*C." + interface0.CTypeAttr

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
	for _, method := range interface0.Methods {
		if strSliceContains(funcs, method.CIdentifier) {
			pMethod(s, method)
		}
	}
}

func pMethod(s *SourceFile, method *mygi.Function) {
	spew.Dump(method)
	s.GoBody.Pn("// wrap for %s", method.CIdentifier)

	instanceParam := method.Parameters.InstanceParameter
	// instanceParam 必须不为空

	instanceParamTpl := newParamPassInTemplate(instanceParam)
	recv := instanceParamTpl.GetVarTypeForGo()

	params := method.Parameters.Parameters
	var paramTpls []*ParamPassInTemplate
	var args []string
	for _, param := range params {
		paramTpl := newParamPassInTemplate(param)
		paramTpls = append(paramTpls, paramTpl)
		args = append(args, paramTpl.GetVarTypeForGo())
	}

	argsJoined := strings.Join(args, ", ")

	retValTpl := newReturnValueTemplate(method.ReturnValue)

	var retTypes []string
	retTypes = append(retTypes, retValTpl.GetTypeForGo())

	retTypesJoined := strings.Join(retTypes, ", ")
	if strings.Contains(retTypesJoined, ",") {
		retTypesJoined = "(" + retTypesJoined + ")"
	}
	s.GoBody.Pn("func (%s) %s (%s) %s {", recv, method.Name(), argsJoined, retTypesJoined)

	// start func body
	instanceParamTpl.WriteDeclaration(s)

	for _, paramTpl := range paramTpls {
		paramTpl.WriteDeclaration(s)
	}

	var exprsInCall []string
	exprsInCall = append(exprsInCall, instanceParamTpl.GetExprInCall())
	for _, paramTpl := range paramTpls {
		exprsInCall = append(exprsInCall, paramTpl.GetExprInCall())
	}

	s.GoBody.Pn("ret0 := C.%s(%s)", method.CIdentifier, strings.Join(exprsInCall, ", "))

	retValTpl.WriteClean(s)

	s.GoBody.Pn("return %s", retValTpl.NormalReturn())

	s.GoBody.Pn("}") // end body
}
