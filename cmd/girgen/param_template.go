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
	if param.Type != nil {
		return newInParamTemplate(param)
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

func (tpl *OutParamTemplate) VarForGo() string {
	return tpl.varForGo
}

func (tpl *OutParamTemplate) TypeForGo() string {
	return tpl.bridge.TypeForGo
}

func (tpl *OutParamTemplate) ExprForC() string {
	return "&" + tpl.varForC
}

func (*OutParamTemplate) ExprForGo() string {
	panic("implement me")
}

func (*OutParamTemplate) ErrExprForGo() string {
	panic("implement me")
}

func (tpl *OutParamTemplate) pBeforeCall(s *SourceFile) {
	panic("implement me")
}

func (tpl *OutParamTemplate) pAfterCall(s *SourceFile) {
	panic("implement me")
}
