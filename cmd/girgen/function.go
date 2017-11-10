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

var goKeywordMap map[string]struct{}

func init() {
	goKeywordMap = make(map[string]struct{})
	for _, kw := range goKeywords {
		goKeywordMap[kw] = struct{}{}
	}
}

func getParamName(in string) string {
	if _, ok := goKeywordMap[in]; ok {
		return in + "_"
	}
	return in
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

func markClosure(fn *gi.FunctionInfo) {
	params := fn.Parameters
	if params != nil {
		for _, param := range params.Parameters {
			if param.ClosureIndex >= 0 {
				params.Parameters[param.ClosureIndex].ClosureForCallbackParam = param
			}
		}
	}
}

func getCFuncArg(p *gi.Parameter) string {
	return p.Type.CType + " " + p.Name
}

func pFunctionWrapper(s *SourceFile, fn *gi.FunctionInfo) {
	returnType := fn.ReturnValue.Type.CType
	instanceParam := fn.Parameters.InstanceParameter

	var typeArgs []string

	if instanceParam != nil {
		typeArgs = append(typeArgs, getCFuncArg(instanceParam))
	}

	// typeArgs 中省略 user data, 将 callback 的类型换成 GClosure*
	// args 中 callback 设置成对应的 CallbackWrapper， user_data 设置成 user_data_for_callback
	// 这样就能和 pFunction 中的实现保持一致了。

	// loop for argsTypes
	for _, param := range fn.Parameters.Parameters {
		if param.ClosureForCallbackParam == nil {
			if param.ClosureIndex >= 0 {
				// param is callback
				closureParam := fn.Parameters.Parameters[param.ClosureIndex]
				arg := closureParam.Name + "_for_" + param.Name
				typeArgs = append(typeArgs, "GClosure* "+arg)
			} else {
				// common
				typeArgs = append(typeArgs, getCFuncArg(param))
			}

		} else {
			// param is user data for callback
			// 省略
		}
	}

	if fn.Throws {
		typeArgs = append(typeArgs, "GError **error")
	}

	argTypesJoined := strings.Join(typeArgs, ", ")
	s.CBody.Pn("static %s _%s(%s) {", returnType, fn.CIdentifier, argTypesJoined)

	var args []string
	if instanceParam != nil {
		args = append(args, instanceParam.Name)
	}
	// loop for args
	for _, param := range fn.Parameters.Parameters {
		if param.ClosureForCallbackParam == nil {
			if param.ClosureIndex >= 0 {
				// param is callback
				args = append(args, param.Type.Name+"Wrapper")
			} else {
				// common
				args = append(args, param.Name)
			}
		} else {
			// param is user data for callback
			arg := param.Name + "_for_" + param.ClosureForCallbackParam.Name
			args = append(args, arg)
		}
	}

	if fn.Throws {
		args = append(args, "error")
	}

	argsJoined := strings.Join(args, ", ")
	s.CBody.Pn("    return %s(%s);", fn.CIdentifier, argsJoined)

	s.CBody.Pn("}")
}

func pFunction(s *SourceFile, fn *gi.FunctionInfo) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("pFunction", fn.CIdentifier)
			panic(err)
		}
	}()
	markLength(fn)
	markClosure(fn)
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

	var hasClosure bool

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
			if param.ClosureForCallbackParam != nil {
				hasClosure = true
				continue
			}

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
		if repo.Namespace.Name == "GLib" {
			s.GoBody.Pn("var err Error")
		} else {
			s.AddGirImport("GLib")
			s.GoBody.Pn("var err glib.Error")
		}
	}

	for _, paramTpl := range paramTpls {
		exprsInCall = append(exprsInCall, paramTpl.ExprForC())
	}
	if fn.Throws {
		exprsInCall = append(exprsInCall, "(**C.GError)(unsafe.Pointer(&err))")
	}

	funcName := fn.CIdentifier
	if hasClosure {
		funcName = "_" + funcName
	}
	call := fmt.Sprintf("C.%s(%s)", funcName, strings.Join(exprsInCall, ", "))
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

	if hasClosure {
		pFunctionWrapper(s, fn)
	}
}
