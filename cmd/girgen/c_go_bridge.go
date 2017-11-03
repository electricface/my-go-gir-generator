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

type ParamTemplate struct {
	VarForC  string
	VarForGo string
	bridge   *CGoBridge
}

func newParamTemplate(param *gi.Parameter) *ParamTemplate {
	tpl := new(ParamTemplate)
	tpl.VarForC = param.Name + "0"
	tpl.VarForGo = param.Name

	// param.Type -> bridge
	tpl.bridge = getBridge(param.Type)
	if tpl.bridge == nil {
		cType, err := gi.ParseCType(param.Type.CType)
		if err != nil {
			panic(err)
		}
		panic(fmt.Errorf("fail to get bridge for type %s,%s", cType.CgoNotation(), param.Type.Name))
	}
	return tpl
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

// go to c
func pParamGo2C(s *SourceFile, t *ParamTemplate) {
	s.GoBody.Pn("\n// Var for Go: %s", t.VarForGo)
	s.GoBody.Pn("// Var for C: %s", t.VarForC)
	s.GoBody.Pn("// Type for Go: %s", t.bridge.TypeForGo)
	s.GoBody.Pn("// Type for C: %s", t.bridge.TypeForC)
	if t.bridge.CvtGo2C != "" {
		s.GoBody.Pn("%s := %s", t.VarForC, t.CvtGo2C())

		if t.bridge.CleanCvtGo2C != "" {
			s.GoBody.Pn("defer %s", t.CleanCvtGo2C())
		}
	}
}

func pParamC2Go(s *SourceFile, t *ParamTemplate) {
	if t.bridge.CvtC2Go != "" {
		s.GoBody.Pn("%s := %s", t.VarForGo, t.CvtC2Go())

		if t.bridge.CleanCvtGo2C != "" {
			s.GoBody.Pn("defer %s", t.CleanCvtC2Go())
		}
	}
}

func getBridge(typ *gi.Type) *CGoBridge {
	typeDef, ns := repo.GetType(typ.Name)
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
			cType, err := gi.ParseCType(typ.CType)
			if err != nil {
				panic(err)
			}

			var isGPointer bool
			if cType.CgoNotation() == "C.gpointer" {
				isGPointer = true
			}

			if cType.NumStar != 1 && !isGPointer {
				panic("assert failed cType.NumStr == 1, ctype is " + typ.CType)
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

	cType, err := gi.ParseCType(typ.CType)
	if err != nil {
		panic(err)
	}
	typeForC := cType.CgoNotation()

	br := getBridgeForIntegerType(typeForC)
	if br != nil {
		return br
	}

	key := typeForC + "," + typ.Name
	return cGoBridgeMap[key]
}

func (tpl *ParamTemplate) VarTypeForGo() string {
	return tpl.VarForGo + " " + tpl.bridge.TypeForGo
}

// go -> c
func (tpl *ParamTemplate) CvtGo2C() string {
	return tpl.replace(tpl.bridge.CvtGo2C)
}

func (tpl *ParamTemplate) CleanCvtGo2C() string {
	return tpl.replace(tpl.bridge.CleanCvtGo2C)
}

func (tpl *ParamTemplate) ExprForC() string {
	return tpl.replace(tpl.bridge.ExprForC)
}

// c -> go
func (tpl *ParamTemplate) CvtC2Go() string {
	return tpl.replace(tpl.bridge.CvtC2Go)
}

func (tpl *ParamTemplate) CleanCvtC2Go() string {
	return tpl.replace(tpl.bridge.CleanCvtC2Go)
}

func (tpl *ParamTemplate) ExprForGo() string {
	return tpl.replace(tpl.bridge.ExprForGo)
}

func (tpl *ParamTemplate) ErrExprForGo() string {
	return tpl.replace(tpl.bridge.ErrExprForGo)
}

func (tpl *ParamTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", tpl.VarForGo,
		"$c", tpl.VarForC)
	return replacer.Replace(in)
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
}

func init() {
	cGoBridgeMap["*C.char,filename"] = cGoBridgeMap["*C.char,utf8"]
	cGoBridgeMap["*C.gchar,filename"] = cGoBridgeMap["*C.gchar,utf8"]
}
