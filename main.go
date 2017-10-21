package main

import (
	"mygi"
	"log"
	"github.com/davecgh/go-spew/spew"
	"fmt"
	"strings"
)

var libCfg *LibConfig

func main() {
	repo, err := mygi.Load("Gio", "2.0")
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
		"g_file_replace":
			pMethod(s, method)
		}
	}
}

var goParamPassInDescMap = map[string]*GoParamPassInDesc{
	// interface
	"*C.GAppInfo": {
		TypeForGo:    "AppInfo",
		TypeForC:     "*C.GAppInfo",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "$g.native()",
	},

	"*C.GFile": {
		TypeForGo:    "File",
		TypeForC:     "*C.GFile",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "$g.native()",
	},

	"*C.GCancellable": {
		TypeForGo:    "GCancellable",
		TypeForC:     "*C.GCancellable",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "$g.native()",
	},

	// boolean
	"C.gboolean": {
		TypeForGo:    "bool",
		TypeForC:     "C.gboolean",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "toGBool($g)",
	},

	// enum
	"C.GFileCreateFlags": {
		TypeForGo: "FileCreateFlags",
		TypeForC: "C.GFileCreateFlags",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "C.GFileCreateFlags($g)",
	},

	// string
	"*C.char": {
		TypeForGo: "string",
		TypeForC: "*C.GAppInfo",
		ConvertExpr: "C.CString($g)",
		ConvertClean:"defer C.free(unsafe.Pointer($c))",
		ExprInCall: "$c",
	},
}

// go参数传入过程的描述
type GoParamPassInDesc struct {
	TypeForGo    string // go
	TypeForC     string // key, c
	ConvertExpr  string
	ConvertClean string
	ExprInCall   string
}

type ParamPassInTemplate struct {
	param *mygi.Parameter
	desc *GoParamPassInDesc
}

func (t *ParamPassInTemplate) replace(in string) string {
	in = strings.Replace(in, "$g", t.GetVarForGo(), -1)
	return strings.Replace(in, "$c", t.GetVarForC(), -1)
}

func (t *ParamPassInTemplate) WriteDeclaration(s *SourceFile) {
	if t.desc.ConvertExpr != "" {
		s.GoBody.Pn("%s := %s", t.GetVarForC(), t.replace(t.desc.ConvertExpr))

		if t.desc.ConvertClean != "" {
			s.GoBody.Pn(t.replace(t.desc.ConvertClean))
		}
	}
}

func (t *ParamPassInTemplate) GetExprInCall() string {
	return t.replace(t.desc.ExprInCall)
}

func (t *ParamPassInTemplate) GetVarForC() string {
	return t.param.Name + "0"
}

func (t *ParamPassInTemplate) GetVarForGo() string {
	return t.param.Name
}

func (t *ParamPassInTemplate) GetVarTypeForGo() string {
	return fmt.Sprintf("%s %s", t.GetVarForGo(), t.desc.TypeForGo )
}

func getGoParamPassInDesc(typeForC string) *GoParamPassInDesc {
	return goParamPassInDescMap[typeForC]
}

func newParamPassInTemplate(param *mygi.Parameter) *ParamPassInTemplate {
	ctype, err := mygi.ParseCType(param.Type.CType)
	if err != nil {
		panic(err)
	}
	typeForC := ctype.CgoNotation()
	passInDesc := getGoParamPassInDesc(typeForC)
	if passInDesc == nil {
		panic("fail to get passInDesc for " + typeForC)
	}

	return &ParamPassInTemplate{
		param: param,
		desc: passInDesc,
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