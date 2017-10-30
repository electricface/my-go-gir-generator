package main

import (
	"fmt"
	"strings"

	"github.com/electricface/my-go-gir-generator/gi"
)

var returnValueDescMap = map[string]*ReturnValueDesc{
	"*C.char": {
		TypeForGo:     "string",
		TypeForC:      "*C.char",
		ReturnExpr:    "C.GoString($c)",
		Clean:         "defer C.g_free(C.gpointer($c))",
		ErrReturnExpr: `""`,
	},
	"*C.gchar": {
		TypeForGo:     "string",
		TypeForC:      "*C.gchar",
		ReturnExpr:    "C.GoString((*C.char)($c))",
		Clean:         "defer C.g_free(C.gpointer($c))",
		ErrReturnExpr: `""`,
	},

	"C.gboolean": {
		TypeForGo:     "bool",
		TypeForC:      "C.gboolean",
		ReturnExpr:    "util.Int2Bool(int($c))/*go:.util*/",
		ErrReturnExpr: "false",
	},

	"C.gdouble": {
		TypeForGo:     "float64",
		TypeForC:      "C.double",
		ReturnExpr:    "float64($c)",
		ErrReturnExpr: "0.0",
	},

	"C.gpointer": {
		TypeForGo:     "unsafe.Pointer",
		TypeForC:      "C.gpointer",
		ReturnExpr:    "unsafe.Pointer($c)",
		ErrReturnExpr: "unsafe.Pointer(nil)",
	},

	// record in other package
	//"*C.GVariant": {
	//	TypeForGo:     "glib.Variant",
	//	TypeForC:      "*C.GVariant",
	//	ReturnExpr:    "glib.WrapVariant(unsafe.Pointer($c))",
	//	ErrReturnExpr: "glib.Variant{}",
	//},

	// class
	//"*C.GSettings": {
	//	TypeForGo:     "Settings",
	//	TypeForC:      "*C.GSettings",
	//	ReturnExpr:    "wrapSettings($c)",
	//	ErrReturnExpr: "Settings{}",
	//},
}

type ReturnValueTemplate struct {
	param *gi.Parameter
	desc  *ReturnValueDesc
}

type ReturnValueDesc struct {
	TypeForC string

	TypeForGo     string
	ReturnExpr    string
	ErrReturnExpr string
	Clean         string
}

func getReturnValueDesc(ty *gi.Type) *ReturnValueDesc {
	// TODO
	typeDef, ns := repo.GetType(ty.Name)
	sameNs := isSameNamespace(ns)
	ns = strings.ToLower(ns)

	if typeDef != nil {
		switch typeDef0 := typeDef.(type) {
		case *gi.ObjectInfo:
			_ = typeDef0
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
			var retExpr string
			if sameNs {
				typeForGo = typeDef.Name()
				if isGPointer {
					retExpr = "Wrap" + typeForGo + "(unsafe.Pointer($c))"
				} else {
					retExpr = "wrap" + typeForGo + "($c)"
				}

			} else {
				typeForGo = ns + "." + typeDef.Name()
				retExpr = ns + ".Wrap" + typeDef.Name() + "(unsafe.Pointer($c))" +
					fmt.Sprintf("/*gir:%s*/", ns)

			}

			return &ReturnValueDesc{
				TypeForGo:     typeForGo,
				TypeForC:      cType.CgoNotation(),
				ReturnExpr:    retExpr,
				ErrReturnExpr: typeForGo + "{}",
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
			var retExpr string
			if sameNs {
				typeForGo = typeDef.Name()
				retExpr = "wrap" + typeForGo + "($c)"
			} else {
				typeForGo = ns + "." + typeDef.Name()
				// 比如 glib.WrapGVariant(unsafe.Pointer(ret0))
				retExpr = ns + ".Wrap" + typeDef.Name() + "(unsafe.Pointer($c))" +
					fmt.Sprintf("/*gir:%s*/", ns)
			}

			return &ReturnValueDesc{
				TypeForGo:     typeForGo,
				TypeForC:      cType.CgoNotation(),
				ReturnExpr:    retExpr,
				ErrReturnExpr: typeForGo + "{}",
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
			var retExpr string
			if sameNs {
				typeForGo = typeDef.Name()
				retExpr = "wrap" + typeForGo + "($c)"
			} else {

				typeForGo = ns + "." + typeDef.Name()
				retExpr = ns + ".Wrap" + typeDef.Name() + "(unsafe.Pointer($c))" +
					fmt.Sprintf("/*gir:%s*/", ns)
			}

			return &ReturnValueDesc{
				TypeForGo:     typeForGo,
				TypeForC:      cType.CgoNotation(),
				ReturnExpr:    retExpr,
				ErrReturnExpr: typeForGo + "{}",
			}
		}
	}

	cType, err := gi.ParseCType(ty.CType)
	if err != nil {
		panic(err)
	}
	typeForC := cType.CgoNotation()

	desc := getReturnValueDescForIntegerType(typeForC)
	if desc != nil {
		// typeForC is integer type
		return desc
	}

	return returnValueDescMap[typeForC]
}

func getReturnValueDescForIntegerType(cgoType string) *ReturnValueDesc {
	typ := strings.TrimPrefix(cgoType, "C.")
	switch typ {
	case "gint", "guint",
		"gint8", "guint8",
		"gint16", "guint16",
		"gint32", "guint32",
		"gint64", "guint64":

		typeForGo := strings.TrimPrefix(typ, "g")

		return &ReturnValueDesc{
			TypeForGo:     typeForGo,
			TypeForC:      cgoType,
			ReturnExpr:    typeForGo + "($c)",
			ErrReturnExpr: "0",
		}

	default:
		return nil
	}
}

func newReturnValueTemplate(param *gi.Parameter) *ReturnValueTemplate {
	desc := getReturnValueDesc(param.Type)
	if desc == nil {
		panic("fail to get returnValueDesc for " + param.Type.CType)
	}

	return &ReturnValueTemplate{
		param: param,
		desc:  desc,
	}
}

func (t *ReturnValueTemplate) replace(in string) string {
	in = strings.Replace(in, "$C", t.desc.TypeForC, -1)
	in = strings.Replace(in, "$g", t.GetVarForGo(), -1)
	return strings.Replace(in, "$c", t.GetVarForC(), -1)
}

func (t *ReturnValueTemplate) GetVarForC() string {
	return "ret0"
}

func (t *ReturnValueTemplate) GetVarForGo() string {
	// 应该用不到
	return "ret"
}

func (t *ReturnValueTemplate) GetTypeForGo() string {
	return t.desc.TypeForGo
}

func (t *ReturnValueTemplate) WriteClean(s *SourceFile) {
	ownership := t.param.TransferOwnership
	if ownership == "none" {
		return
	}

	if ownership == "full" {
		if t.desc.Clean != "" {
			s.GoBody.Pn(t.replace(t.desc.Clean))
		}
	}
}

func (t *ReturnValueTemplate) NormalReturn() string {
	return t.replace(t.desc.ReturnExpr)
}

func (t *ReturnValueTemplate) ErrorReturn() string {
	return t.replace(t.desc.ErrReturnExpr)
}
