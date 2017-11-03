package main

import (
	"fmt"
	"github.com/electricface/my-go-gir-generator/gi"
	"strings"
)

type ParamTemplate interface {
	VarForGo() string
	TypeForGo() string

	pGo2C(s *SourceFile)
	ExprForC() string

	pC2Go(s *SourceFile)
	ExprForGo() string
	ErrExprForGo() string
}

type SimpleParamTemplate struct {
	varForC  string
	varForGo string
	bridge   *CGoBridge
}

func newParamTemplate(param *gi.Parameter) *SimpleParamTemplate {
	tpl := new(SimpleParamTemplate)
	tpl.varForC = param.Name + "0"
	tpl.varForGo = param.Name

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

func (tpl *SimpleParamTemplate) VarForGo() string {
	return tpl.varForGo
}

func (tpl *SimpleParamTemplate) TypeForGo() string {
	return tpl.bridge.TypeForGo
}

func (tpl *SimpleParamTemplate) ExprForC() string {
	return tpl.replace(tpl.bridge.ExprForC)
}

func (tpl *SimpleParamTemplate) ExprForGo() string {
	return tpl.replace(tpl.bridge.ExprForGo)
}

func (tpl *SimpleParamTemplate) ErrExprForGo() string {
	return tpl.replace(tpl.bridge.ErrExprForGo)
}

func (tpl *SimpleParamTemplate) pGo2C(s *SourceFile) {
	s.GoBody.Pn("\n// Var for Go: %s", tpl.varForGo)
	s.GoBody.Pn("// Var for C: %s", tpl.varForC)
	s.GoBody.Pn("// Type for Go: %s", tpl.bridge.TypeForGo)
	s.GoBody.Pn("// Type for C: %s", tpl.bridge.TypeForC)
	if tpl.bridge.CvtGo2C != "" {
		s.GoBody.Pn("%s := %s", tpl.varForC, tpl.CvtGo2C())

		if tpl.bridge.CleanCvtGo2C != "" {
			s.GoBody.Pn("defer %s", tpl.CleanCvtGo2C())
		}
	}
}

func (tpl *SimpleParamTemplate) pC2Go(s *SourceFile) {
	if tpl.bridge.CvtC2Go != "" {
		s.GoBody.Pn("%s := %s", tpl.varForGo, tpl.CvtC2Go())

		if tpl.bridge.CleanCvtGo2C != "" {
			s.GoBody.Pn("defer %s", tpl.CleanCvtC2Go())
		}
	}
}

// go -> c
func (tpl *SimpleParamTemplate) CvtGo2C() string {
	return tpl.replace(tpl.bridge.CvtGo2C)
}

func (tpl *SimpleParamTemplate) CleanCvtGo2C() string {
	return tpl.replace(tpl.bridge.CleanCvtGo2C)
}

// c -> go
func (tpl *SimpleParamTemplate) CvtC2Go() string {
	return tpl.replace(tpl.bridge.CvtC2Go)
}

func (tpl *SimpleParamTemplate) CleanCvtC2Go() string {
	return tpl.replace(tpl.bridge.CleanCvtC2Go)
}

func (tpl *SimpleParamTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", tpl.varForGo,
		"$c", tpl.varForC)
	return replacer.Replace(in)
}
