package main

import (
	"mygi"
	"strings"
)

var returnValueDescMap = map[string]*ReturnValueDesc{
	"*C.char": {
		TypeForGo: "string",
		TypeForC: "*C.char",
		ReturnExpr: "C.GoString($c)",
		ErrReturnExpr: `""`,
	},
	"*C.gchar": {
		TypeForGo: "string",
		TypeForC: "*C.gchar",
		ReturnExpr: "C.GoString((*C.char)($c))",
		ErrReturnExpr: `""`,
	},

	"*C.GVariant":{
		TypeForGo: "glib.Variant",
		TypeForC: "*C.GVariant",
		ReturnExpr: "glib.WrapVariant(unsafe.Pointer($c))",
		ErrReturnExpr: "glib.Variant{}",
	},
	"*C.GFileOutputStream": {
		TypeForGo: "FileOutputStream",
		TypeForC: "*C.GFileOutputStream",
		ReturnExpr: "wrapFileOutputStream($c)",
		ErrReturnExpr: "FileOutputStream{}",
	},
}

type ReturnValueTemplate struct {
	param *mygi.Parameter
	desc *ReturnValueDesc
}

type ReturnValueDesc struct {
	TypeForC string

	TypeForGo string
	ReturnExpr string
	ErrReturnExpr string
}

func getReturnValueDesc(ty *mygi.Type) *ReturnValueDesc {
	cType, err := mygi.ParseCType(ty.CType)
	if err != nil {
		panic(err)
	}
	typeForC := cType.CgoNotation()
	return returnValueDescMap[typeForC]
}

func newReturnValueTemplate(param *mygi.Parameter) *ReturnValueTemplate {
	desc := getReturnValueDesc(param.Type)
	if desc == nil {
		panic("fail to get returnValueDesc for " + param.Type.CType)
	}

	return &ReturnValueTemplate{
		param: param,
		desc: desc,
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

func (t *ReturnValueTemplate) NormalReturn() string {
	return t.replace(t.desc.ReturnExpr)
}

func (t *ReturnValueTemplate) ErrorReturn() string {
	return t.replace(t.desc.ErrReturnExpr)
}
