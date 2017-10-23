package main

import (
	"mygi"
	"strings"
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
		ReturnExpr:    "mygiutil.Int2Bool(int($c))",
		ErrReturnExpr: "false",
	},

	// gdouble
	"C.gdouble": {
		TypeForGo:     "float64",
		TypeForC:      "C.double",
		ReturnExpr:    "float64($c)",
		ErrReturnExpr: "0.0",
	},

	"*C.GVariant": {
		TypeForGo:     "glib.Variant",
		TypeForC:      "*C.GVariant",
		ReturnExpr:    "glib.WrapVariant(unsafe.Pointer($c))",
		ErrReturnExpr: "glib.Variant{}",
	},
	"*C.GFileOutputStream": {
		TypeForGo:     "FileOutputStream",
		TypeForC:      "*C.GFileOutputStream",
		ReturnExpr:    "wrapFileOutputStream($c)",
		ErrReturnExpr: "FileOutputStream{}",
	},
	// class
	"*C.GSettings": {
		TypeForGo:     "Settings",
		TypeForC:      "*C.GSettings",
		ReturnExpr:    "wrapSettings($c)",
		ErrReturnExpr: "Settings{}",
	},
}

type ReturnValueTemplate struct {
	param *mygi.Parameter
	desc  *ReturnValueDesc
}

type ReturnValueDesc struct {
	TypeForC string

	TypeForGo     string
	ReturnExpr    string
	ErrReturnExpr string
	Clean         string
}

func getReturnValueDesc(ty *mygi.Type) *ReturnValueDesc {
	cType, err := mygi.ParseCType(ty.CType)
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

func newReturnValueTemplate(param *mygi.Parameter) *ReturnValueTemplate {
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
