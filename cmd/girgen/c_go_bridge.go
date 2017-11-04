package main

import (
	"fmt"
	"github.com/electricface/my-go-gir-generator/gi"
	"strings"
)

type CGoBridge struct {
	TypeForC  string
	TypeForGo string

	// go -> c
	CvtGo2C      string
	CleanCvtGo2C string
	ExprForC     string

	// c -> go
	CvtC2Go      string
	CleanCvtC2Go string
	ExprForGo    string
	ErrExprForGo string
}

func getBridgeForIntegerType(cgoType string) *CGoBridge {
	typ := strings.TrimPrefix(cgoType, "C.")
	switch typ {
	case "int", "uint",
		"gint", "guint",
		"gint8", "guint8",
		"gint16", "guint16",
		"gint32", "guint32",
		"gint64", "guint64":

		typeForGo := strings.TrimPrefix(typ, "g")

		return &CGoBridge{
			TypeForGo: typeForGo,
			TypeForC:  cgoType,
			ExprForC:  "$C($g)",
			ExprForGo: "$G($c)",
		}

	default:
		return nil
	}
}

func isSameNamespace(ns string) bool {
	return ns == repo.Namespace.Name
}

func getBridge(typeName string, cType *gi.CType) *CGoBridge {
	typeDef, ns := repo.GetType(typeName)
	sameNs := isSameNamespace(ns)
	nsLower := strings.ToLower(ns)

	if typeDef != nil {
		_, isEnum := typeDef.(*gi.EnumInfo)
		_, isStruct := typeDef.(*gi.StructInfo)
		_, isObject := typeDef.(*gi.ObjectInfo)
		_, isInterface := typeDef.(*gi.InterfaceInfo)

		if isEnum {
			var typeForGo string
			if sameNs {
				typeForGo = typeDef.Name()
			} else {
				typeForGo = nsLower + "." + typeDef.Name()
			}

			return &CGoBridge{
				TypeForGo: typeForGo,
				TypeForC:  typeDef.CType().CgoNotation(),

				ExprForC: "$C($g)",

				ExprForGo: typeForGo + "($c)",
			}
		}

		if isStruct || isObject || isInterface {
			var isGPointer bool
			if cType.CgoNotation() == "C.gpointer" {
				isGPointer = true
			}

			if cType.NumStar != 1 && !isGPointer {
				panic("assert failed cType.NumStr == 1, ctype is " + cType.CgoNotation())
			}

			var typeForGo string
			var exprForC string
			var exprForGo string
			if sameNs {
				typeForGo = typeDef.Name()
				exprForC = "$g.native()"
				if isGPointer {
					exprForGo = "Wrap" + typeForGo + "(unsafe.Pointer($c))"
				} else {
					exprForGo = "wrap" + typeForGo + "($c)"
				}
			} else {
				typeForGo = nsLower + "." + typeDef.Name()
				// 不能使用 native 方法了
				// 比如 (*C.GFile)(file.Ptr)
				exprForC = "($C)($g.Ptr)"
				//typeForGo = ns + "." + typeDef.Name()
				exprForGo = nsLower + ".Wrap" + typeDef.Name() + "(unsafe.Pointer($c))" +
					fmt.Sprintf("/*gir:%s*/", ns)
			}
			if isGPointer {
				exprForC = fmt.Sprintf("C.gpointer(%s)", exprForC)
			}

			return &CGoBridge{
				TypeForGo: typeForGo,
				TypeForC:  cType.CgoNotation(),

				ExprForC:     exprForC,
				ExprForGo:    exprForGo,
				ErrExprForGo: typeForGo + "{}",
			}
		}
	}

	typeForC := cType.CgoNotation()

	br := getBridgeForIntegerType(typeForC)
	if br != nil {
		return br
	}

	key := typeForC + "," + typeName
	return cGoBridgeMap[key]
}

var cGoBridgeMap = map[string]*CGoBridge{
	"*C.char,utf8": {
		TypeForC:  "*C.char",
		TypeForGo: "string",

		CvtGo2C:      "C.CString($g)",
		CleanCvtGo2C: "C.free(unsafe.Pointer($c)) /*ch:<stdlib.h>*/",
		ExprForC:     "$c",

		CvtC2Go:      "C.GoString($c)",
		CleanCvtC2Go: "C.g_free(C.gpointer($c))",
		ExprForGo:    "$g",
		ErrExprForGo: `""`,
	},

	"*C.gchar,utf8": {
		TypeForC:  "*C.gchar",
		TypeForGo: "string",

		CvtGo2C:      "(*C.gchar)(C.CString($g))",
		CleanCvtGo2C: "C.free(unsafe.Pointer($c)) /*ch:<stdlib.h>*/",
		ExprForC:     "$c",

		CvtC2Go:      "C.GoString( (*C.char)($c) )",
		CleanCvtC2Go: "C.g_free(C.gpointer($c))",
		ExprForGo:    "$g",
		ErrExprForGo: `""`,
	},

	"C.gboolean,gboolean": {
		TypeForC:  "C.gboolean",
		TypeForGo: "bool",

		ExprForC: "$C(util.Bool2Int($g)) /*go:.util*/",

		ExprForGo:    "util.Int2Bool(int($c)) /*go:.util*/",
		ErrExprForGo: "false",
	},

	"C.gdouble,gdouble": {
		TypeForGo: "float64",
		TypeForC:  "C.gdouble",

		ExprForC:     "$C($g)",
		ExprForGo:    "$G($c)",
		ErrExprForGo: "0.0",
	},

	"C.gfloat,gfloat": {
		TypeForGo: "float32",
		TypeForC:  "C.gfloat",

		ExprForC:     "$C($g)",
		ExprForGo:    "$G($c)",
		ErrExprForGo: "0.0",
	},

	"C.gpointer,gpointer": {
		TypeForGo: "unsafe.Pointer",
		TypeForC:  "C.gpointer",

		ExprForC:     "$C($g)",
		ExprForGo:    "$G($c)",
		ErrExprForGo: "unsafe.Pointer(nil)",
	},

	"C.guchar,guint8": {
		TypeForGo: "byte",
		TypeForC:  "C.guchar",

		ExprForC:     "$C($g)",
		ExprForGo:    "$G($c)",
		ErrExprForGo: "0",
	},

	"C.gsize,gsize": {
		TypeForGo: "uint",
		TypeForC:  "C.gsize",

		ExprForC:     "$C($g)",
		ExprForGo:    "$G($c)",
		ErrExprForGo: "0",
	},
}

func init() {
	cGoBridgeMap["*C.char,filename"] = cGoBridgeMap["*C.char,utf8"]
	cGoBridgeMap["*C.gchar,filename"] = cGoBridgeMap["*C.gchar,utf8"]
}
