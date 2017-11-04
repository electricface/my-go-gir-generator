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
	tpl.varForC = param.Name + "0"
	tpl.varForGo = param.Name

	// param.Type -> bridge
	cType, err := gi.ParseCType(param.Type.CType)
	if err != nil {
		panic(err)
	}
	tpl.bridge = getBridge(param.Type.Name, cType)
	if tpl.bridge == nil {
		panic(fmt.Errorf("fail to get bridge for type %s,%s", cType.CgoNotation(), param.Type.Name))
	}
	return tpl
}

type InParamTemplate struct {
	varForC  string
	varForGo string
	bridge   *CGoBridge
}

func (tpl *InParamTemplate) VarForGo() string {
	return tpl.varForGo
}

func (tpl *InParamTemplate) TypeForGo() string {
	return tpl.bridge.TypeForGo
}

func (tpl *InParamTemplate) ExprForC() string {
	return tpl.replace(tpl.bridge.ExprForC)
}

func (tpl *InParamTemplate) ExprForGo() string {
	return tpl.replace(tpl.bridge.ExprForGo)
}

func (tpl *InParamTemplate) ErrExprForGo() string {
	return tpl.replace(tpl.bridge.ErrExprForGo)
}

func (tpl *InParamTemplate) pBeforeCall(s *SourceFile) {
	s.GoBody.Pn("\n// Var for Go: %s", tpl.varForGo)
	s.GoBody.Pn("// Var for C: %s", tpl.varForC)
	s.GoBody.Pn("// Type for Go: %s", tpl.bridge.TypeForGo)
	s.GoBody.Pn("// Type for C: %s", tpl.bridge.TypeForC)
	if tpl.bridge.CvtGo2C != "" {
		s.GoBody.Pn("%s := %s", tpl.varForC, tpl.replace(tpl.bridge.CvtGo2C))
	}
}

func (tpl *InParamTemplate) pAfterCall(s *SourceFile) {
	if tpl.bridge.CvtGo2C != "" && tpl.bridge.CleanCvtGo2C != "" {
		s.GoBody.Pn("%s", tpl.replace(tpl.bridge.CleanCvtGo2C))
	}
}

func (tpl *InParamTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", tpl.varForGo,
		"$c", tpl.varForC)
	return replacer.Replace(in)
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
	tpl.varForGo = param.Name

	// param.Type -> bridge
	cType, err := gi.ParseCType(param.Type.CType)
	if err != nil {
		panic(err)
	}

	realCType := cType.Elem()

	tpl.bridge = getBridge(param.Type.Name, realCType)
	if tpl.bridge == nil {
		panic(fmt.Errorf("fail to get bridge for type %s,%s",
			realCType.CgoNotation(), param.Type.Name))
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
		s.GoBody.Pn("%s", tpl.replace(tpl.bridge.CvtC2Go))
	}
	//s.GoBody.Pn("%s := %s", tpl.varForGo, tpl.replace(tpl.bridge.ExprForGo))

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
