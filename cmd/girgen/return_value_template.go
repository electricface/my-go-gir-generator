package main

import (
	"fmt"
	"github.com/electricface/my-go-gir-generator/gi"
	"strings"
)

type ReturnValueTemplate interface {
	TypeForGo() string
	pAfterCall(s *SourceFile)

	ExprForGo() string
	ErrExprForGo() string
}

func newReturnValueTemplate(param *gi.Parameter) ReturnValueTemplate {
	if param.Type != nil {
		return newSimpleReturnValueTemplate(param)
	}
	if param.Array != nil {
		return newArrayReturnValueTemplate(param)
	}
	return nil
}

// return value
// not array
type SimpleReturnValueTemplate struct {
	varForGo          string
	varForC           string
	bridge            *CGoBridge
	transferOwnership string
}

func newSimpleReturnValueTemplate(param *gi.Parameter) *SimpleReturnValueTemplate {
	tpl := new(SimpleReturnValueTemplate)
	tpl.varForC = param.Name + "0"
	tpl.varForGo = param.Name
	tpl.transferOwnership = param.TransferOwnership

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

func (tpl *SimpleReturnValueTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", tpl.varForGo,
		"$c", tpl.varForC)
	return replacer.Replace(in)
}

func (tpl *SimpleReturnValueTemplate) TypeForGo() string {
	return tpl.bridge.TypeForGo
}

func (tpl *SimpleReturnValueTemplate) pAfterCall(s *SourceFile) {
	if tpl.bridge.CvtC2Go != "" {
		s.GoBody.Pn("%s := %s", tpl.varForGo, tpl.replace(tpl.bridge.CvtC2Go))
	}

	if tpl.transferOwnership == "full" &&
		tpl.bridge.CvtC2Go != "" && tpl.bridge.CleanCvtC2Go != "" {
		s.GoBody.Pn("%s", tpl.replace(tpl.bridge.CleanCvtC2Go))
	}
}

func (tpl *SimpleReturnValueTemplate) ExprForGo() string {
	return tpl.replace(tpl.bridge.ExprForGo)
}

func (tpl *SimpleReturnValueTemplate) ErrExprForGo() string {
	return tpl.replace(tpl.bridge.ErrExprForGo)
}

// return value
// array
type ArrayReturnValueTemplate struct {
	varForGo          string
	varForC           string
	bridge            *CGoBridge
	array             *gi.ArrayType
	elemCType         *gi.CType
	transferOwnership string
}

func newArrayReturnValueTemplate(param *gi.Parameter) *ArrayReturnValueTemplate {
	array := param.Array

	tpl := new(ArrayReturnValueTemplate)
	tpl.varForGo = param.Name
	tpl.varForC = param.Name + "0"
	tpl.transferOwnership = param.TransferOwnership
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

func (tpl *ArrayReturnValueTemplate) TypeForGo() string {
	return "[]" + tpl.bridge.TypeForGo
}

func (tpl *ArrayReturnValueTemplate) ExprForGo() string {
	return tpl.varForGo
}

func (tpl *ArrayReturnValueTemplate) ErrExprForGo() string {
	return "nil"
}

func (tpl *ArrayReturnValueTemplate) replace(in string) string {
	replacer := strings.NewReplacer("$C", tpl.bridge.TypeForC,
		"$G", tpl.bridge.TypeForGo,
		"$g", "elemG",
		"$c", "elem")
	return replacer.Replace(in)
}

func (tpl *ArrayReturnValueTemplate) getCSliceLengthExpr() string {
	if tpl.array.LengthParameter != nil {
		return fmt.Sprintf("int(%s)", tpl.array.LengthParameter.Name+"0")
	}

	if tpl.array.ZeroTerminated {
		var elemInstance string
		if tpl.elemCType.NumStar > 0 {
			elemInstance = "uintptr(0)"
		} else {
			// ex. C.GType(0)
			elemInstance = tpl.elemCType.CgoNotation() + "(0)"
		}
		return fmt.Sprintf("\n util.GetZeroTermArrayLen(unsafe.Pointer(%s), unsafe.Sizeof(%s)) /*go:.util*/",
			tpl.varForC, elemInstance)
	}
	return "0 /*TODO*/"
}

func (tpl *ArrayReturnValueTemplate) pAfterCall(s *SourceFile) {
	cSlice := tpl.varForC + "Slice"
	s.GoBody.Pn("var %s []%s", cSlice, tpl.bridge.TypeForC)

	//s.GoBody.Pn("%sLength := %s", cSlice, tpl.getCSliceLengthExpr())
	s.GoBody.Pn("util.SetSliceDataLen(unsafe.Pointer(&%s), unsafe.Pointer(%s), %s) /*go:.util*/",
		cSlice, tpl.varForC, tpl.getCSliceLengthExpr())
	s.GoBody.Pn("%s := make([]%s, len(%s))", tpl.varForGo, tpl.bridge.TypeForGo, cSlice)
	s.GoBody.Pn("for idx, elem := range %s {", cSlice)

	if tpl.bridge.CvtC2Go != "" {
		s.GoBody.Pn("    elemG := %s", tpl.replace(tpl.bridge.CvtC2Go))
	}
	s.GoBody.Pn("    %s[idx] = %s", tpl.varForGo, tpl.replace(tpl.bridge.ExprForGo))

	if tpl.transferOwnership == "full" &&
		tpl.bridge.CvtC2Go != "" && tpl.bridge.CleanCvtC2Go != "" {
		s.GoBody.Pn("    %s", tpl.replace(tpl.bridge.CleanCvtC2Go))
	}

	s.GoBody.Pn("}") // end for

	// free container
	if tpl.transferOwnership == "full" || tpl.transferOwnership == "container" {
		s.GoBody.Pn("C.g_free(C.gpointer(%s))", tpl.varForC)
	}
}
