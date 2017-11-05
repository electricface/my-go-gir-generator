package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	repo, err = gi.Load(cfg.Namespace, cfg.Version)
	if err != nil {
		log.Fatal(err)
	}

	types := repo.GetTypes()
	log.Print(len(types))
	pkg := strings.ToLower(repo.Namespace.Name)
	sourceFile := getSourceFile(repo, pkg)

	for _, genFileCfg := range cfg.GenFiles {
		typeDef, ns := repo.GetType(genFileCfg.Type)
		if typeDef == nil {
			panic("fail to get type for " + genFileCfg.Type)
		}
		if ns != cfg.Namespace {
			panic("assert failed ns == cfg.Namespace")
		}

		switch td := typeDef.(type) {
		case *gi.StructInfo:
			pStruct(sourceFile, td, genFileCfg.Funcs)
		case *gi.InterfaceInfo:
			pInterface(sourceFile, td, genFileCfg.Funcs)
		case *gi.ObjectInfo:
			pObject(sourceFile, td, genFileCfg.Funcs)
		}
	}

	for _, fn := range repo.Namespace.Functions {
		if strSliceContains(cfg.Funcs, fn.CIdentifier) {
			pFunction(sourceFile, fn)
		}
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

	//repo.GetType()
	//
	//for name, type0 := range types {
	//	log.Printf("%s -> %T\n", name, type0)
	//}
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

func getOutputFileBaseName(cfg *GenFileConfig) string {
	var name string
	if cfg.Filename != "" {
		name = cfg.Filename
	} else {
		name = camel2Snake(cfg.Type)
	}
	return name + "_auto.go"
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

func getVarTypeForGo(tpl ParamTemplate) string {
	return tpl.VarForGo() + " " + tpl.TypeForGo()
}

func IsFuncReturnVoid(retVal *gi.Parameter) bool {
	if retVal.Type != nil {
		return retVal.Type.Name == "none"
	}
	// else is array?
	if retVal.Array == nil {
		panic("assert failed retVal.Array != nil")
	}
	return false
}

func markLength(fn *gi.FunctionInfo) {
	params := fn.Parameters
	if params != nil {
		for _, param := range params.Parameters {
			if param.Array != nil {
				lenIdx := param.Array.LengthIndex
				if lenIdx >= 0 {
					params.Parameters[lenIdx].LengthForParameter = param
					param.Array.LengthParameter = params.Parameters[lenIdx]
				}
			}
		}
	}

	retVal := fn.ReturnValue
	if !IsFuncReturnVoid(retVal) {
		if retVal.Array != nil {
			lenIdx := retVal.Array.LengthIndex
			if lenIdx >= 0 {
				params.Parameters[lenIdx].LengthForParameter = retVal
				retVal.Array.LengthParameter = params.Parameters[lenIdx]
			}
		}
	}
}

func pFunction(s *SourceFile, fn *gi.FunctionInfo) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("pFunction", fn.CIdentifier)
			panic(err)
		}
	}()
	markLength(fn)
	s.GoBody.Pn("// %s is a wrapper around %s().", fn.Name(), fn.CIdentifier)

	var receiver string
	var args []string
	var retTypes []string
	var instanceParamTpl ParamTemplate
	var paramTpls []ParamTemplate
	var retValTpl ReturnValueTemplate
	var retVals []string
	var errRetVals []string

	// return order: return value, param-out, error

	if !IsFuncReturnVoid(fn.ReturnValue) {
		fn.ReturnValue.Name = "ret"
		retValTpl = newReturnValueTemplate(fn.ReturnValue)
		if retValTpl == nil {
			panic("newReturnValueTemplate failed")
		}
		retTypes = append(retTypes, retValTpl.TypeForGo())
		retVals = append(retVals, retValTpl.ExprForGo())
		errRetVals = append(errRetVals, retValTpl.ErrExprForGo())
	}

	if fn.Parameters != nil {
		instanceParam := fn.Parameters.InstanceParameter
		if instanceParam != nil {
			instanceParamTpl = newParamTemplate(instanceParam)
			if instanceParamTpl == nil {
				panic("newParamTemplate failed for instance param " + instanceParam.Name)
			}
			receiver = "(" + getVarTypeForGo(instanceParamTpl) + ")"
		}

		for _, param := range fn.Parameters.Parameters {
			tpl := newParamTemplate(param)
			if tpl == nil {
				panic("newParamTemplate failed for param " + param.Name)
			}

			paramTpls = append(paramTpls, tpl)

			if param.Direction == "" {
				// direction in
				if param.LengthForParameter == nil {
					args = append(args, getVarTypeForGo(tpl))
				}
			} else if param.Direction == "out" {

				if param.LengthForParameter == nil {
					retTypes = append(retTypes, tpl.TypeForGo())
					retVals = append(retVals, tpl.ExprForGo())
					errRetVals = append(errRetVals, tpl.ErrExprForGo())
				}

			} else if param.Direction == "inout" {
				panic("todo")
			} else {
				panic("invalid param direction")
			}

		}
	}

	if fn.Throws {
		retTypes = append(retTypes, "error")
	}

	argsJoined := strings.Join(args, ", ")
	retTypesJoined := strings.Join(retTypes, ", ")
	if strings.Contains(retTypesJoined, ",") {
		retTypesJoined = "(" + retTypesJoined + ")"
	}
	s.GoBody.Pn("func %s %s (%s) %s {", receiver, fn.Name(), argsJoined, retTypesJoined)

	// start func body
	var exprsInCall []string
	if instanceParamTpl != nil {
		instanceParamTpl.pBeforeCall(s)
		exprsInCall = append(exprsInCall, instanceParamTpl.ExprForC())
	}

	for _, paramTpl := range paramTpls {
		paramTpl.pBeforeCall(s)
	}

	if fn.Throws {
		s.AddGirImport("GLib")
		s.GoBody.Pn("var err glib.Error")
	}

	for _, paramTpl := range paramTpls {
		exprsInCall = append(exprsInCall, paramTpl.ExprForC())
	}
	if fn.Throws {
		exprsInCall = append(exprsInCall, "(**C.GError)(unsafe.Pointer(&err))")
	}

	call := fmt.Sprintf("C.%s(%s)", fn.CIdentifier, strings.Join(exprsInCall, ", "))
	if retValTpl != nil {
		s.GoBody.P("ret0 := ")
	}
	s.GoBody.Pn(call)

	// after call
	if instanceParamTpl != nil {
		instanceParamTpl.pAfterCall(s)
	}
	for _, paramTpl := range paramTpls {
		paramTpl.pAfterCall(s)
	}

	if retValTpl != nil {
		retValTpl.pAfterCall(s)
	}

	retValsJoined := strings.Join(retVals, ", ")
	if fn.Throws {
		errRetValsJoined := strings.Join(errRetVals, ", ")

		s.GoBody.Pn("if err.Ptr != nil {")
		s.GoBody.Pn("defer err.Free()")
		s.GoBody.Pn("return %s, err.GoValue()", errRetValsJoined)
		s.GoBody.Pn("}") // end if

		s.GoBody.Pn("return %s, nil", retValsJoined)

	} else if len(retVals) > 0 {
		s.GoBody.Pn("return %s", retValsJoined)
	}
	s.GoBody.Pn("}") // end func
}
