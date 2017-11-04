package main

import (
	"fmt"
	"strings"

	"github.com/electricface/my-go-gir-generator/gi"
)

type ParamTemplate interface {
	VarForGo() string
	TypeForGo() string

	ExprForC() string
	pGo2CBeforeCall(s *SourceFile)
	pGo2CAfterCall(s *SourceFile)

	ExprForGo() string
	ErrExprForGo() string

	pC2GoBeforeCall(s *SourceFile)
	pC2GoAfterCall(s *SourceFile)
}

func newParamTemplate(param *gi.Parameter) ParamTemplate {
	if param.Type != nil {
		return newSimpleParamTemplate(param)
	}
	if param.Array != nil {
		return newArrayParamTemplate(param)
	}
	return nil
}

func newSimpleParamTemplate(param *gi.Parameter) *SimpleParamTemplate {
	tpl := new(SimpleParamTemplate)
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

type SimpleParamTemplate struct {
	varForC  string
	varForGo string
	bridge   *CGoBridge
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

func (tpl *SimpleParamTemplate) pGo2CBeforeCall(s *SourceFile) {
	s.GoBody.Pn("\n// Var for Go: %s", tpl.varForGo)
	s.GoBody.Pn("// Var for C: %s", tpl.varForC)
	s.GoBody.Pn("// Type for Go: %s", tpl.bridge.TypeForGo)
	s.GoBody.Pn("// Type for C: %s", tpl.bridge.TypeForC)
	if tpl.bridge.CvtGo2C != "" {
		s.GoBody.Pn("%s := %s", tpl.varForC, tpl.replace(tpl.bridge.CvtGo2C))
	}
}

func (tpl *SimpleParamTemplate) pGo2CAfterCall(s *SourceFile) {
	if tpl.bridge.CvtGo2C != "" && tpl.bridge.CleanCvtGo2C != "" {
		s.GoBody.Pn("%s", tpl.replace(tpl.bridge.CleanCvtGo2C))
	}
}

func (tpl *SimpleParamTemplate) pC2GoBeforeCall(s *SourceFile) {
}

func (tpl *SimpleParamTemplate) pC2GoAfterCall(s *SourceFile) {
	if tpl.bridge.CvtC2Go != "" {
		s.GoBody.Pn("%s := %s", tpl.varForGo, tpl.replace(tpl.bridge.CvtC2Go))
	}

	if tpl.bridge.CvtC2Go != "" && tpl.bridge.CleanCvtC2Go != "" {
		s.GoBody.Pn("%s", tpl.replace(tpl.bridge.CleanCvtC2Go))
	}
}

func (tpl *SimpleParamTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", tpl.varForGo,
		"$c", tpl.varForC)
	return replacer.Replace(in)
}

type ArrayParamTemplate struct {
	varForGo  string
	varForC   string
	bridge    *CGoBridge
	array     *gi.ArrayType
	elemCType *gi.CType
}

func newArrayParamTemplate(param *gi.Parameter) *ArrayParamTemplate {
	array := param.Array

	tpl := new(ArrayParamTemplate)
	tpl.varForGo = param.Name
	tpl.varForC = param.Name + "0"
	tpl.array = array

	arrayCType, err := gi.ParseCType(array.CType)
	if err != nil {
		panic(err)
	}
	elemCType := arrayCType.Elem()
	tpl.elemCType = elemCType

	tpl.bridge = getBridge(array.ElemType.Name, elemCType)
	return tpl
}

func (tpl *ArrayParamTemplate) VarForGo() string {
	return tpl.varForGo
}

func (tpl *ArrayParamTemplate) TypeForGo() string {
	return "[]" + tpl.bridge.TypeForGo
}

func (tpl *ArrayParamTemplate) ExprForC() string {
	// TODO:
	return tpl.replace(tpl.bridge.ExprForC)
}

func (tpl *ArrayParamTemplate) ExprForGo() string {
	return tpl.varForGo
}

func (tpl *ArrayParamTemplate) ErrExprForGo() string {
	return "nil"
}

func (tpl *ArrayParamTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", "elemG",
		"$c", "elem")
	return replacer.Replace(in)
}

func (tpl *ArrayParamTemplate) getCSliceLengthExpr() string {
	var elemInstance string
	if tpl.elemCType.NumStar > 0 {
		elemInstance = "uintptr(0)"
	} else {
		// ex. C.GType(0)
		elemInstance = tpl.elemCType.CgoNotation() + "(0)"
	}

	if tpl.array.ZeroTerminated {
		return fmt.Sprintf("util.GetZeroTermArrayLen(unsafe.Pointer(%s), unsafe.Sizeof(%s)) /*go:.util*/",
			tpl.varForC, elemInstance)
	}
	return "0 /*TODO*/"
}

func (tpl *ArrayParamTemplate) pC2GoBeforeCall(s *SourceFile) {

}

func (tpl *ArrayParamTemplate) pC2GoAfterCall(s *SourceFile) {
	cSlice := tpl.varForC + "Slice"
	s.GoBody.Pn("var %s []%s", cSlice, tpl.bridge.TypeForC)

	s.GoBody.Pn("%sLength := %s", cSlice, tpl.getCSliceLengthExpr())
	s.GoBody.Pn("util.SetSliceDataLen(unsafe.Pointer(&%s), unsafe.Pointer(%s), %sLength) /*go:.util*/",
		cSlice, tpl.varForC, cSlice)
	s.GoBody.Pn("%s := make([]%s, len(%s))", tpl.varForGo, tpl.bridge.TypeForGo, cSlice)
	s.GoBody.Pn("for idx, elem := range %s {", cSlice)

	if tpl.bridge.CvtC2Go != "" {
		s.GoBody.Pn("    elemG := %s", tpl.replace(tpl.bridge.CvtC2Go))
		if tpl.bridge.CleanCvtC2Go != "" {
			s.GoBody.Pn("    defer %s", tpl.replace(tpl.bridge.CleanCvtC2Go))
		}
	}
	s.GoBody.Pn("    %s[idx] = %s", tpl.varForGo, tpl.replace(tpl.bridge.ExprForGo))

	s.GoBody.Pn("}") // end for
}

func (tpl *ArrayParamTemplate) pGo2CBeforeCall(s *SourceFile) {
	// TODO:
}

func (tpl *ArrayParamTemplate) pGo2CAfterCall(s *SourceFile) {
	// TODO:
}
