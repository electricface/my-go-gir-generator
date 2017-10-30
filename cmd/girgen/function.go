package main

import (
//"fmt"
//"github.com/electricface/my-go-gir-generator/gi"
//"strings"
//"./typeconverter"
)

var goKeywords = []string{
	"break", "default", "func", "interface", "select",
	"case", "defer", "go", "map", "struct",
	"chan", "else", "goto", "package", "switch",
	"const", "fallthrough", "if", "range", "type",
	"continue", "for", "import", "return", "var",
}

//func getParamName(p *gi.Parameter) string {
//    name := strings.TrimPrefix(p.Name, "@")
//    if strSliceContains(goKeywords, name) {
//        return strings.ToUpper(name)
//    }
//    return name
//}
//
//func getParamGoType(p *gi.Parameter, paramCvtMap map[string]*typeconverter.TypeConverter) string {
//    cvt := paramCvtMap[p.Name]
//    return cvt.GoType
//}
//
//func getParamCParamName(p *gi.Parameter) string {
//    return "C_" + getParamName(p)
//}
//
//func getParamCParamCgoType(p *gi.Parameter, paramCvtMap map[string]*typeconverter.TypeConverter) string {
//    cvt := paramCvtMap[p.Name]
//    return cvt.CgoType
//}
//
//func getParamInCCall(p *gi.Parameter) string {
//    if p.Direction == "out" {
//        return "&C_" + getParamName(p)
//    }
//    return "C_" + getParamName(p)
//}
//
//func getParamReturnName(p *gi.Parameter) string {
//    return "R_" + getParamName(p)
//}
//
//func newErrorParameter() *gi.Parameter {
//    return &gi.Parameter{
//        Name: "err",
//        TransferOwnership: "all",
//        Direction: "out",
//        Type: gi.Type{
//            Name: "GError",
//            CType: "GError**",
//        },
//    }
//}
//
//
//func printFunction(f *gi.Function, funcName string) {
//    if !f.Introspectable {
//        fmt.Println("// not introspectable")
//        return
//    }
//    //if libCfg.IsFunctionInBlacklist(f.CIdentifier) {
//    //    fmt.Println("// function in blacklist")
//    //    return
//    //}
//
//    //var paramCvtMap map[string]*TypeConverter
//    paramCvtMap := make(map[string]*typeconverter.TypeConverter)
//
//    fparams := f.Parameters
//    var params []*gi.Parameter
//    instanceParam := fparams.InstanceParameter
//    if !instanceParam.IsZero() {
//        instanceParam.Name = "@this"
//        params = append(params, &instanceParam)
//    }
//    fparamsCount := len(fparams.Parameters)
//    // handle array length param
//    for i := 0; i < fparamsCount; i++ {
//        p := &fparams.Parameters[i]
//        //fmt.Printf("[%d] p is %#v\n", i, p)
//        params = append(params, p)
//        arr := p.Array
//        if arr.IsZero() || arr.LengthIndex < 0 {
//            continue
//        }
//        fparams.Parameters[arr.LengthIndex].LengthForParameter = p
//        //fmt.Println("set length for param", arr.LengthIndex, p)
//    }
//    //for i,param := range fparams.Parameters {
//        //fmt.Println("check length for param", i,param.Name, param.LengthForParameter)
//    //}
//
//    // add return
//    retVal := &f.ReturnValue
//    if retVal.Type.Name != "none" {
//        retVal.Name = "@ret"
//        params = append(params, retVal)
//    }
//
//    // add error
//    if f.Throws {
//        params = append(params, newErrorParameter())
//    }
//
//    for _, p := range params {
//        isRetVal := p.Name == "@ret"
//        paramCvtMap[p.Name] = typeconverter.NewTypeConverter(isRetVal, p.Direction, &p.Type, &p.Array)
//    }
//
//    // c call args
//    var cArgs []*gi.Parameter
//    for _, p := range params {
//        if p.Name != "@ret" {
//            cArgs = append(cArgs, p)
//        }
//    }
//    printParamters("cArgs", cArgs)
//
//    // arg in
//    // may has @this
//    // no @ret
//    var cParamsFromGo []*gi.Parameter
//    for _, p := range params {
//        if (p.Direction == "" || p.Direction == "in"  || p.Direction == "inout") &&
//            p.Name != "@ret" {
//            cParamsFromGo = append(cParamsFromGo, p)
//        }
//    }
//    printParamters("cParamsFromGo", cParamsFromGo)
//
//    // arg out
//    // has @ret
//    var cParamsToGo []*gi.Parameter
//    for _, p := range params {
//        if (p.Direction == "out" || p.Direction == "inout") ||
//            p.Name == "@ret" {
//            cParamsToGo = append(cParamsToGo, p)
//        }
//    }
//    printParamters("cParamsToGo / goReturns", cParamsToGo)
//
//    var goArgs []*gi.Parameter
//    for _, p := range cParamsFromGo {
//        if p.LengthForParameter == nil {
//            goArgs = append(goArgs, p)
//        }
//    }
//    printParamters("goArgs", goArgs)
//    pFunctionHeader(funcName, goArgs, cParamsToGo, paramCvtMap)
//
//    pCArgDeclarations(cArgs, paramCvtMap)
//    pCParamsFromGoConversions(cParamsFromGo, paramCvtMap)
//
//    pCCall(f.CIdentifier, retVal, cArgs)
//
//    pReturnDeclarations(cParamsToGo, paramCvtMap)
//    pCParamsToGoConversions(cParamsToGo, paramCvtMap)
//
//    pFunctionReturn(cParamsToGo)
//    fmt.Println("}")
//}
//
//// for debug
//func printParamters(name string, params []*gi.Parameter) {
//    return
//    fmt.Println("// -> params", name)
//    for _,p := range params {
//        fmt.Printf("// %s %#v\n", p.Name, p.Type)
//    }
//}
//
//func pFunctionHeader(name string, goArgs []*gi.Parameter,
//    goReturns []*gi.Parameter, paramCvtMap map[string]*typeconverter.TypeConverter) {
//
//    fmt.Printf("func ")
//    if len(goArgs) > 0 {
//        if goArgs[0].Name == "@this" {
//            fmt.Printf("(this %s) ", getParamGoType(goArgs[0], paramCvtMap))
//            goArgs = goArgs[1:]
//        }
//    }
//    fmt.Printf("%s(", name)
//    // go args
//    var goArgNameTypes []string
//    for _, param := range goArgs {
//        goArgNameTypes = append(goArgNameTypes,
//            getParamName(param) + " " + getParamGoType(param, paramCvtMap))
//    }
//    fmt.Printf("%s) ", strings.Join(goArgNameTypes, ", "))
//
//    // return type
//    var goReturnTypes []string
//    for _, param := range goReturns {
//        goReturnTypes = append(goReturnTypes, getParamGoType(param, paramCvtMap))
//    }
//    ret := strings.Join(goReturnTypes, ", ")
//    if len(goReturnTypes) > 1 {
//        ret = "(" + ret + ")"
//    }
//    fmt.Printf("%s {\n", ret)
//}
//
//func pCArgDeclarations(cArgs []*gi.Parameter, paramCvtMap map[string]*typeconverter.TypeConverter) {
//    for _, param := range cArgs {
//        fmt.Printf("\tvar %s %s\n", getParamCParamName(param), getParamCParamCgoType(param, paramCvtMap) )
//    }
//}
//
//// go -> cgo
//func pCParamsFromGoConversions(cParamsFromGo []*gi.Parameter,
//    paramCvtMap map[string]*typeconverter.TypeConverter) {
//
//    for _, param := range cParamsFromGo {
//        cvt := paramCvtMap[param.Name]
//        fmt.Printf("\t// go2cgo: %s,%s\n", cvt.GoType, cvt.CgoType)
//        cvt.ConvertGo2Cgo(getParamCParamName(param), getParamName(param))
//    }
//}
//
//// cgo -> go
//func pCParamsToGoConversions(cParamsToGo []*gi.Parameter,
//    paramCvtMap map[string]*typeconverter.TypeConverter) {
//    for _, param := range cParamsToGo {
//        cvt := paramCvtMap[param.Name]
//        fmt.Printf("\t// cgo2go: %s,%s\n", cvt.CgoType, cvt.GoType)
//        cvt.ConvertCgo2Go(getParamReturnName(param), getParamCParamName(param))
//    }
//}
//
//func pReturnDeclarations(goReturns []*gi.Parameter,
//    paramCvtMap map[string]*typeconverter.TypeConverter) {
//    for _, param := range goReturns {
//        fmt.Printf("\tvar %s %s\n", getParamReturnName(param), getParamGoType(param, paramCvtMap) )
//    }
//}
//
//func pCCall(name string, retVal *gi.Parameter, cArgs []*gi.Parameter) {
//    if retVal.Type.Name != "none" {
//        fmt.Printf("\t%s := ", getParamInCCall(retVal))
//    } else {
//        fmt.Printf("\t")
//    }
//    var cCallArgs []string
//    for _, param := range cArgs {
//        cCallArgs = append(cCallArgs, getParamInCCall(param))
//    }
//    fmt.Printf("C.%s(%s)\n", name, strings.Join(cCallArgs, ", "))
//}
//
//func pFunctionReturn(goReturns []*gi.Parameter) {
//    fmt.Printf("\treturn")
//    if len(goReturns) == 0 {
//        fmt.Println()
//        return
//    }
//    var rets []string
//    for _, param := range goReturns {
//        rets = append(rets, getParamReturnName(param))
//    }
//    fmt.Printf(" %s\n", strings.Join(rets, ", ") )
//}
//
//func pMethod(f *gi.Function) {
//    fmt.Println("// method", f.CIdentifier)
//    printFunction(f, f.Name())
//}
//
//func pFunction(f *gi.Function) {
//    fmt.Println("// function", f.CIdentifier)
//    printFunction(f, f.Name())
//}
//
//func pConstructor(f *gi.Function, className string) {
//    fmt.Println("// constructor", f.CIdentifier)
//    newName := strings.Replace( f.Name(),"New", "New" + className , 1)
//    // New -> NewAppInfo
//    // NewFromFilename -> NewAppInfoFromFilename
//    printFunction(f, newName)
//}
