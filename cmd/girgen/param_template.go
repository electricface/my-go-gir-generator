package main

import (
	"fmt"
	"strings"

	"github.com/electricface/my-go-gir-generator/gi"
)

type ParamTemplate interface {
	VarForGo() string
	TypeForGo() string
	pBeforeCall(s *SourceFile)
	pAfterCall(s *SourceFile)
	ExprForC() string
	ExprForGo() string
	ErrExprForGo() string
}

func newParamTemplate(param *gi.Parameter) ParamTemplate {
	// direction in, out, inout
	if param.Direction == "" {
		// direction in
		if param.Type != nil {
			return newInParamTemplate(param)
		}
		if param.Array != nil {
			return newInArrayParamTemplate(param)
		}
	} else if param.Direction == "out" {
		if param.Type != nil {
			return newOutParamTemplate(param)
		}
	}

	return nil
}

// param
// direction = in
func newInParamTemplate(param *gi.Parameter) *InParamTemplate {
	tpl := new(InParamTemplate)
	tpl.varForGo = getParamName(param.Name)
	tpl.varForC = param.Name + "0"

	// param.Type -> bridge
	cType, err := gi.ParseCType(param.Type.CType)
	if err != nil {
		panic(err)
	}
	tpl.bridge, err = getBridge(param.Type.Name, cType)
	if err != nil {
		panic(err)
	}

	if param.LengthForParameter != nil {
		tpl.lengthForParameter = param.LengthForParameter.Name
	}

	if param.ClosureParam != nil {
		// param is callback
		tpl.isClosure = true
		scope := param.Scope
		if scope == "" {
			scope = "call" // default scope is call
		}
		tpl.closureScope = scope
	}

	return tpl
}

type InParamTemplate struct {
	varForC            string
	varForGo           string
	bridge             *CGoBridge
	lengthForParameter string
	isClosure          bool
	closureScope       string
}

func (tpl *InParamTemplate) VarForGo() string {
	return tpl.varForGo
}

func (tpl *InParamTemplate) TypeForGo() string {
	return tpl.bridge.TypeForGo
}

func (tpl *InParamTemplate) ExprForC() string {
	if tpl.lengthForParameter != "" {
		return fmt.Sprintf("%s(len(%s))", tpl.bridge.TypeForC, tpl.lengthForParameter)
	}

	return tpl.replace(tpl.bridge.ExprForC)
}

func (tpl *InParamTemplate) ExprForGo() string {
	panic("call the function should not be called")
}

func (tpl *InParamTemplate) ErrExprForGo() string {
	panic("call the function should not be called")
}

func (tpl *InParamTemplate) pBeforeCall(s *SourceFile) {
	if tpl.bridge.CvtGo2C != "" {
		s.GoBody.Pn("%s := %s", tpl.varForC, tpl.replace(tpl.bridge.CvtGo2C))
	}
}

func (tpl *InParamTemplate) pAfterCall(s *SourceFile) {
	if tpl.bridge.CvtGo2C != "" && tpl.bridge.CleanCvtGo2C != "" {

		if tpl.isClosure {
			if tpl.closureScope == "call" {
				s.GoBody.Pn("%s", tpl.replace(tpl.bridge.CleanCvtGo2C))
			}

		} else {
			s.GoBody.Pn("%s", tpl.replace(tpl.bridge.CleanCvtGo2C))
		}
	}
}

func (tpl *InParamTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", tpl.varForGo,
		"$c", tpl.varForC)
	return replacer.Replace(in)
}

type InArrayParamTemplate struct {
	varForC   string
	varForGo  string
	bridge    *CGoBridge
	array     *gi.ArrayType
	elemCType *gi.CType
}

func newInArrayParamTemplate(param *gi.Parameter) *InArrayParamTemplate {
	array := param.Array
	tpl := new(InArrayParamTemplate)
	tpl.varForC = param.Name + "0"
	tpl.varForGo = getParamName(param.Name)
	tpl.array = array

	arrayCType, err := gi.ParseCType(array.CType)
	if err != nil {
		panic(err)
	}

	var elemCType *gi.CType
	if array.CType == "gconstpointer" {
		if array.ElemType.Name == "guint8" {
			elemCType, _ = gi.ParseCType("guint8")
		} else {
			panic("todo")
		}

	} else {
		elemCType = arrayCType.Elem()
	}
	tpl.elemCType = elemCType
	tpl.bridge, err = getBridge(array.ElemType.Name, elemCType)
	if err != nil {
		panic(err)
	}
	return tpl
}

func (tpl *InArrayParamTemplate) VarForGo() string {
	return tpl.varForGo
}

func (tpl *InArrayParamTemplate) TypeForGo() string {
	return "[]" + tpl.bridge.TypeForGo
}

func (tpl *InArrayParamTemplate) pBeforeCall(s *SourceFile) {
	s.GoBody.Pn("%s := make([]%s, len(%s))", tpl.varForC, tpl.bridge.TypeForC, tpl.varForGo)
	s.GoBody.Pn("for idx, elemG := range %s {", tpl.varForGo)

	if tpl.bridge.CvtGo2C != "" {
		s.GoBody.Pn("    elem := %s", tpl.replace(tpl.bridge.CvtGo2C))
	}
	s.GoBody.Pn("    %s[idx] = %s", tpl.varForC, tpl.replace(tpl.bridge.ExprForC))
	s.GoBody.Pn("}") // end for

	s.GoBody.Pn("var %sPtr *%s", tpl.varForC, tpl.bridge.TypeForC)
	s.GoBody.Pn("if len(%s) > 0 {", tpl.varForC)
	s.GoBody.Pn("    %sPtr = &%s[0]", tpl.varForC, tpl.varForC)
	s.GoBody.Pn("}")
}

func (tpl *InArrayParamTemplate) pAfterCall(s *SourceFile) {
	if tpl.bridge.CvtGo2C != "" && tpl.bridge.CleanCvtGo2C != "" {
		s.GoBody.Pn("for _, elem := range %s {", tpl.varForC)
		s.GoBody.Pn("    %s", tpl.replace(tpl.bridge.CleanCvtGo2C))
		s.GoBody.Pn("}")
	}
}

func (tpl *InArrayParamTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", "elemG",
		"$c", "elem")
	return replacer.Replace(in)
}

func (tpl *InArrayParamTemplate) ExprForC() string {
	return tpl.varForC + "Ptr"
}

func (tpl *InArrayParamTemplate) ExprForGo() string {
	panic("call the function should not be called")
}

func (*InArrayParamTemplate) ErrExprForGo() string {
	panic("call the function should not be called")
}

// is param
// direction out
type OutParamTemplate struct {
	varForGo string
	varForC  string
	bridge   *CGoBridge
}

func newOutParamTemplate(param *gi.Parameter) *OutParamTemplate {
	tpl := new(OutParamTemplate)
	tpl.varForC = param.Name + "0"
	tpl.varForGo = getParamName(param.Name)

	// param.Type -> bridge
	cType, err := gi.ParseCType(param.Type.CType)
	if err != nil {
		panic(err)
	}

	realCType := cType.Elem()

	tpl.bridge, err = getBridge(param.Type.Name, realCType)
	if err != nil {
		panic(err)
	}
	return tpl
}

func (tpl *OutParamTemplate) VarForGo() string {
	return tpl.varForGo
}

func (tpl *OutParamTemplate) TypeForGo() string {
	// maybe hide
	return tpl.bridge.TypeForGo
}

func (tpl *OutParamTemplate) ExprForC() string {
	return "&" + tpl.varForC
}

func (tpl *OutParamTemplate) ExprForGo() string {
	return tpl.replace(tpl.bridge.ExprForGo)
}

func (tpl *OutParamTemplate) ErrExprForGo() string {
	return tpl.replace(tpl.bridge.ErrExprForGo)
}

func (tpl *OutParamTemplate) pBeforeCall(s *SourceFile) {
	s.GoBody.Pn("var %s %s", tpl.varForC, tpl.bridge.TypeForC)
}

func (tpl *OutParamTemplate) pAfterCall(s *SourceFile) {

	if tpl.bridge.CvtC2Go != "" {
		s.GoBody.Pn("%s := %s", tpl.varForGo, tpl.replace(tpl.bridge.CvtC2Go))
	}

	if tpl.bridge.CvtC2Go != "" && tpl.bridge.CleanCvtC2Go != "" {
		s.GoBody.Pn("defer %s", tpl.replace(tpl.bridge.CleanCvtC2Go))
	}
}

func (tpl *OutParamTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", tpl.varForGo,
		"$c", tpl.varForC)
	return replacer.Replace(in)
}
