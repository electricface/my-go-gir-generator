package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/electricface/my-go-gir-generator/gi"
)

var goKeywords = []string{
	"break", "default", "func", "interface", "select",
	"case", "defer", "go", "map", "struct",
	"chan", "else", "goto", "package", "switch",
	"const", "fallthrough", "if", "range", "type",
	"continue", "for", "import", "return", "var",
}

func getVarTypeForGo(tpl ParamTemplate) string {
	return tpl.VarForGo() + " " + tpl.TypeForGo()
}

func IsFuncReturnVoid(retVal *gi.Parameter) bool {
	if retVal.Type != nil {
		return retVal.Type.Name == "none"
	}
	// else is array?
	if retVal.Array == nil {
		panic("assert failed retVal.Array != nil")
	}
	return false
}

func markLength(fn *gi.FunctionInfo) {
	params := fn.Parameters
	if params != nil {
		for _, param := range params.Parameters {
			if param.Array != nil {
				lenIdx := param.Array.LengthIndex
				if lenIdx >= 0 {
					params.Parameters[lenIdx].LengthForParameter = param
					param.Array.LengthParameter = params.Parameters[lenIdx]
				}
			}
		}
	}

	retVal := fn.ReturnValue
	if !IsFuncReturnVoid(retVal) {
		if retVal.Array != nil {
			lenIdx := retVal.Array.LengthIndex
			if lenIdx >= 0 {
				params.Parameters[lenIdx].LengthForParameter = retVal
				retVal.Array.LengthParameter = params.Parameters[lenIdx]
			}
		}
	}
}

func pFunction(s *SourceFile, fn *gi.FunctionInfo) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("pFunction", fn.CIdentifier)
			panic(err)
		}
	}()
	markLength(fn)
	s.GoBody.Pn("// %s is a wrapper around %s().", fn.Name(), fn.CIdentifier)

	var receiver string
	var args []string
	var retTypes []string
	var instanceParamTpl ParamTemplate
	var paramTpls []ParamTemplate
	var retValTpl ReturnValueTemplate
	var retVals []string
	var errRetVals []string

	// return order: return value, param-out, error

	if !IsFuncReturnVoid(fn.ReturnValue) {
		fn.ReturnValue.Name = "ret"
		retValTpl = newReturnValueTemplate(fn.ReturnValue)
		if retValTpl == nil {
			panic("newReturnValueTemplate failed")
		}
		retTypes = append(retTypes, retValTpl.TypeForGo())
		retVals = append(retVals, retValTpl.ExprForGo())
		errRetVals = append(errRetVals, retValTpl.ErrExprForGo())
	}

	if fn.Parameters != nil {
		instanceParam := fn.Parameters.InstanceParameter
		if instanceParam != nil {
			instanceParamTpl = newParamTemplate(instanceParam)
			if instanceParamTpl == nil {
				panic("newParamTemplate failed for instance param " + instanceParam.Name)
			}
			receiver = "(" + getVarTypeForGo(instanceParamTpl) + ")"
		}

		for _, param := range fn.Parameters.Parameters {
			tpl := newParamTemplate(param)
			if tpl == nil {
				panic("newParamTemplate failed for param " + param.Name)
			}

			paramTpls = append(paramTpls, tpl)

			if param.Direction == "" {
				// direction in
				if param.LengthForParameter == nil {
					args = append(args, getVarTypeForGo(tpl))
				}
			} else if param.Direction == "out" {

				if param.LengthForParameter == nil {
					retTypes = append(retTypes, tpl.TypeForGo())
					retVals = append(retVals, tpl.ExprForGo())
					errRetVals = append(errRetVals, tpl.ErrExprForGo())
				}

			} else if param.Direction == "inout" {
				panic("todo")
			} else {
				panic("invalid param direction")
			}

		}
	}

	if fn.Throws {
		retTypes = append(retTypes, "error")
	}

	argsJoined := strings.Join(args, ", ")
	retTypesJoined := strings.Join(retTypes, ", ")
	if strings.Contains(retTypesJoined, ",") {
		retTypesJoined = "(" + retTypesJoined + ")"
	}
	s.GoBody.Pn("func %s %s (%s) %s {", receiver, fn.Name(), argsJoined, retTypesJoined)

	// start func body
	var exprsInCall []string
	if instanceParamTpl != nil {
		instanceParamTpl.pBeforeCall(s)
		exprsInCall = append(exprsInCall, instanceParamTpl.ExprForC())
	}

	for _, paramTpl := range paramTpls {
		paramTpl.pBeforeCall(s)
	}

	if fn.Throws {
		s.AddGirImport("GLib")
		s.GoBody.Pn("var err glib.Error")
	}

	for _, paramTpl := range paramTpls {
		exprsInCall = append(exprsInCall, paramTpl.ExprForC())
	}
	if fn.Throws {
		exprsInCall = append(exprsInCall, "(**C.GError)(unsafe.Pointer(&err))")
	}

	call := fmt.Sprintf("C.%s(%s)", fn.CIdentifier, strings.Join(exprsInCall, ", "))
	if retValTpl != nil {
		s.GoBody.P("ret0 := ")
	}
	s.GoBody.Pn(call)

	// after call
	if instanceParamTpl != nil {
		instanceParamTpl.pAfterCall(s)
	}
	for _, paramTpl := range paramTpls {
		paramTpl.pAfterCall(s)
	}

	if retValTpl != nil {
		retValTpl.pAfterCall(s)
	}

	retValsJoined := strings.Join(retVals, ", ")
	if fn.Throws {
		errRetValsJoined := strings.Join(errRetVals, ", ")

		s.GoBody.Pn("if err.Ptr != nil {")
		s.GoBody.Pn("defer err.Free()")
		s.GoBody.Pn("return %s, err.GoValue()", errRetValsJoined)
		s.GoBody.Pn("}") // end if

		s.GoBody.Pn("return %s, nil", retValsJoined)

	} else if len(retVals) > 0 {
		s.GoBody.Pn("return %s", retValsJoined)
	}
	s.GoBody.Pn("}") // end func
}
