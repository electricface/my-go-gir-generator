package main

import (
	"fmt"
	"strings"

	"github.com/electricface/my-go-gir-generator/gi"
)

var goParamPassInDescMap = map[string]*GoParamPassInDesc{
	// interface
	//"*C.GAppInfo": {
	//	TypeForGo:    "AppInfo",
	//	TypeForC:     "*C.GAppInfo",
	//	ConvertExpr:  "",
	//	ConvertClean: "",
	//	ExprInCall:   "$g.native()",
	//},
	//

	// class
	//"*C.GCancellable": {
	//	TypeForGo:    "Cancellable",
	//	TypeForC:     "*C.GCancellable",
	//	ConvertExpr:  "",
	//	ConvertClean: "",
	//	ExprInCall:   "$g.native()",
	//},

	// enum
	//"C.GFileCreateFlags": {
	//	TypeForGo: "FileCreateFlags",
	//	TypeForC: "C.GFileCreateFlags",
	//	ConvertExpr:  "",
	//	ConvertClean: "",
	//	ExprInCall:   "C.GFileCreateFlags($g)",
	//},

	// boolean
	"C.gboolean": {
		TypeForGo:    "bool",
		TypeForC:     "C.gboolean",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "$C(util.Bool2Int($g)) /*go:.util*/",
	},

	"C.gdouble": {
		TypeForGo:    "float64",
		TypeForC:     "C.gdouble",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "$C($g)",
	},

	"C.gfloat": {
		TypeForGo:    "float32",
		TypeForC:     "C.gfloat",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "$C($g)",
	},

	"C.gpointer": {
		TypeForGo:    "unsafe.Pointer",
		TypeForC:     "C.gpointer",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "$C($g)",
	},

	"C.guchar": {
		TypeForGo:    "byte",
		TypeForC:     "C.guchar",
		ConvertExpr:  "",
		ConvertClean: "",
		ExprInCall:   "$C($g)",
	},

	// string
	"*C.char": {
		TypeForGo:    "string",
		TypeForC:     "*C.char",
		ConvertExpr:  "C.CString($g)",
		ConvertClean: "defer C.free(unsafe.Pointer($c)) /*ch:<stdlib.h>*/",
		ExprInCall:   "$c",
	},

	"*C.gchar": {
		TypeForGo:    "string",
		TypeForC:     "*C.gchar",
		ConvertExpr:  "(*C.gchar)(C.CString($g))",
		ConvertClean: "defer C.free(unsafe.Pointer($c)) /*ch:<stdlib.h>*/",
		ExprInCall:   "$c",
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
	param *gi.Parameter
	desc  *GoParamPassInDesc
}

func (t *ParamPassInTemplate) replace(in string) string {
	in = strings.Replace(in, "$C", t.desc.TypeForC, -1)
	in = strings.Replace(in, "$g", t.GetVarForGo(), -1)
	return strings.Replace(in, "$c", t.GetVarForC(), -1)
}

func (t *ParamPassInTemplate) WriteDeclaration(s *SourceFile) {
	s.GoBody.Pn("\n// Var for Go: %s", t.GetVarForGo())
	s.GoBody.Pn("// Var for C: %s", t.GetVarForC())
	s.GoBody.Pn("// Type for Go: %s", t.replace(t.desc.TypeForGo))
	s.GoBody.Pn("// Type for C: %s", t.replace(t.desc.TypeForC))
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
	return fmt.Sprintf("%s %s", t.GetVarForGo(), t.desc.TypeForGo)
}

func isSameNamespace(ns string) bool {
	return ns == repo.Namespace.Name
}

func getGoParamPassInDescForIntegerType(cgoType string) *GoParamPassInDesc {
	typ := strings.TrimPrefix(cgoType, "C.")
	switch typ {
	case "gint", "guint",
		"gint8", "guint8",
		"gint16", "guint16",
		"gint32", "guint32",
		"gint64", "guint64":

		typeForGo := strings.TrimPrefix(typ, "g")

		return &GoParamPassInDesc{
			TypeForGo:    typeForGo,
			TypeForC:     cgoType,
			ConvertExpr:  "",
			ConvertClean: "",
			ExprInCall:   "$C($g)",
		}

	default:
		return nil
	}
}

func getGoParamPassInDesc(ty *gi.Type) *GoParamPassInDesc {
	// TODO
	typeDef, ns := repo.GetType(ty.Name)
	sameNs := isSameNamespace(ns)
	//if typeDef == nil {
	//	panic("failed to get type define for " + ty.Name)
	//}
	if typeDef != nil {
		switch typeDef0 := typeDef.(type) {
		case *gi.EnumInfo:
			_ = typeDef0
			var typeForGo string
			if sameNs {
				typeForGo = typeDef.Name()
			} else {
				typeForGo = strings.ToLower(ns) + "." + typeDef.Name()
			}

			return &GoParamPassInDesc{
				TypeForGo:  typeForGo,
				TypeForC:   typeDef.CType().CgoNotation(),
				ExprInCall: "$C($g)",
			}

		case *gi.StructInfo:
			cType, err := gi.ParseCType(ty.CType)
			if err != nil {
				panic(err)
			}

			if cType.NumStar != 1 {
				panic("assert failed cType.NumStr == 1")
			}

			var typeForGo string
			var exprInCall string
			if sameNs {
				typeForGo = typeDef.Name()
				exprInCall = "$g.native()"
			} else {
				typeForGo = strings.ToLower(ns) + "." + typeDef.Name()
				exprInCall = "($C)($g.Ptr)"
			}

			return &GoParamPassInDesc{
				TypeForGo:  typeForGo,
				TypeForC:   cType.CgoNotation(),
				ExprInCall: exprInCall,
			}

		case *gi.InterfaceInfo:
			cType, err := gi.ParseCType(ty.CType)
			if err != nil {
				panic(err)
			}

			if cType.NumStar != 1 {
				panic("assert failed cType.NumStr == 1")
			}

			var typeForGo string
			var exprInCall string
			if sameNs {
				typeForGo = typeDef.Name()
				exprInCall = "$g.native()"
			} else {
				typeForGo = strings.ToLower(ns) + "." + typeDef.Name()
				// 不能使用 native 方法了
				// 比如 (*C.GFile)(file.Ptr)
				exprInCall = "($C)($g.Ptr)"
			}

			return &GoParamPassInDesc{
				TypeForGo:  typeForGo,
				TypeForC:   cType.CgoNotation(),
				ExprInCall: exprInCall,
			}

		case *gi.ObjectInfo:
			cType, err := gi.ParseCType(ty.CType)
			if err != nil {
				panic(err)
			}

			var isGPointer bool
			if cType.CgoNotation() == "C.gpointer" {
				isGPointer = true
			}

			if cType.NumStar != 1 && !isGPointer {
				panic("assert failed cType.NumStr == 1, ctype is " + ty.CType)
			}

			var typeForGo string
			var exprInCall string
			if sameNs {
				typeForGo = typeDef.Name()
				exprInCall = "$g.native()"
			} else {
				typeForGo = strings.ToLower(ns) + "." + typeDef.Name()
				// 不能使用 native 方法了
				// 比如 (*C.GFile)(file.Ptr)
				exprInCall = "($C)($g.Ptr)"
			}
			if isGPointer {
				exprInCall = fmt.Sprintf("C.gpointer(%s)", exprInCall)
			}

			return &GoParamPassInDesc{
				TypeForGo:  typeForGo,
				TypeForC:   cType.CgoNotation(),
				ExprInCall: exprInCall,
			}
		}
	}

	cType, err := gi.ParseCType(ty.CType)
	if err != nil {
		panic(err)
	}
	typeForC := cType.CgoNotation()

	desc := getGoParamPassInDescForIntegerType(typeForC)
	if desc != nil {
		return desc
	}

	return goParamPassInDescMap[typeForC]
}

func newParamPassInTemplate(param *gi.Parameter) *ParamPassInTemplate {
	passInDesc := getGoParamPassInDesc(param.Type)
	if passInDesc == nil {
		panic("fail to get passInDesc for " + param.Type.CType)
	}

	return &ParamPassInTemplate{
		param: param,
		desc:  passInDesc,
	}
}
