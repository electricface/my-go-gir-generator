package main

import (
	"log"
	"strings"

	"mygi"
	"github.com/davecgh/go-spew/spew"
)

var libCfg *LibConfig
var repo *mygi.Repository

func main() {
	var err error
	repo, err = mygi.Load("Gio", "2.0")
	if err != nil {
		log.Fatal(err)
	}

	//types := repo.GetTypes()
	//log.Print(len(types))
	//
	//for name, type0 := range types {
	//	log.Printf("%s -> %T\n", name, type0)
	//}

	interfaces := repo.Namespace.Interfaces
	for _, interface0 := range interfaces {
		if interface0.Name() == "File" {
			sourceFile := NewSourceFile("gio")
			pInterface(sourceFile, interface0)
			//sourceFile.Print()
			sourceFile.Save("out/appinfo.go")
		}
	}

	classes := repo.Namespace.Classes
	for _, class := range classes {
		if class.Name() == "Settings" {
			sourceFile := NewSourceFile("gio")
			pClass(sourceFile, class)
			//sourceFile.Print()
			sourceFile.Save("out/settings.go")
		}
	}
}

func pClass(s *SourceFile, class *mygi.Class) {
	name := class.Name()
	s.GoBody.Pn("// class %s", name)

	s.GoBody.Pn("type %s struct {", name )
	s.GoBody.Pn("Ptr unsafe.Pointer")
	s.GoBody.Pn("}")

	cPtrType := "*C." + class.CTypeAttr

	// method native
	s.GoBody.Pn("func (v %s) native() %s {", name, cPtrType)
	s.GoBody.Pn("return (%s)(v.Ptr)", cPtrType)
	s.GoBody.Pn("}")

	// method wrapXXX
	s.GoBody.Pn("func wrap%s(p %s) %s {", name, cPtrType, name )
	s.GoBody.Pn("return %s{unsafe.Pointer(p)}", name)
	s.GoBody.Pn("}")

	// method WrapXXX
	s.GoBody.Pn("func Wrap%s(p unsafe.Pointer) %s {", name, name )
	s.GoBody.Pn("return %s{p}", name)
	s.GoBody.Pn("}")

	// methods
	for _, method := range class.Methods {
		switch method.CIdentifier {
		case "g_app_info_get_id",
			"g_app_info_set_as_last_used_for_type",
			"g_file_replace",
			"g_settings_get_value":
			pMethod(s, method)
		}
	}
}

func pInterface(s *SourceFile, interface0 *mygi.Interface) {
	name := interface0.Name()
	s.GoBody.Pn("// interface %s", name)

	s.GoBody.Pn("type %s struct {", name )
	s.GoBody.Pn("Ptr unsafe.Pointer")
	s.GoBody.Pn("}")

	cPtrType := "*C." + interface0.CTypeAttr

	// method native
	s.GoBody.Pn("func (v %s) native() %s {", name, cPtrType)
	s.GoBody.Pn("return (%s)(v.Ptr)", cPtrType)
	s.GoBody.Pn("}")

	// method wrapXXX
	s.GoBody.Pn("func wrap%s(p %s) %s {", name, cPtrType, name )
	s.GoBody.Pn("return %s{unsafe.Pointer(p)}", name)
	s.GoBody.Pn("}")

	// method WrapXXX
	s.GoBody.Pn("func Wrap%s(p unsafe.Pointer) %s {", name, name )
	s.GoBody.Pn("return %s{p}", name)
	s.GoBody.Pn("}")

	// methods
	for _, method := range interface0.Methods {
		switch method.CIdentifier {
		case "g_app_info_get_id",
		"g_app_info_set_as_last_used_for_type",
		"g_file_replace",
		"g_settings_get_value":
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
	rets := ""
	s.GoBody.Pn("func (%s) %s (%s) %s {", recv, method.Name(), argsJoined, rets)

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

	s.GoBody.Pn("ret := C.%s(%s)", method.CIdentifier, strings.Join(exprsInCall,", ") )

	s.GoBody.Pn("}") // end body
}