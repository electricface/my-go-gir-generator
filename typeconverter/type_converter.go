package typeconverter

import (
    "fmt"
    "mygi"
    "strings"
    "bytes"
)

var namespace string

func SetNamespace(ns string) {
    namespace = ns
}

type TypeConverter struct {
    IsReturnValue bool
    Direction string
    Type *mygi.Type
    Array *mygi.Array

    IsArrayType bool
    IsGList bool
    ElemCgoType string
    ElemGoType string

    CgoType string
    GoType string
    CType string

    //Go2Cgo string
    //Cgo2Go string
}

func NewTypeConverter(isRetVal bool,direction string, ty *mygi.Type, arr *mygi.Array) *TypeConverter {
    cvt := &TypeConverter{
        IsReturnValue: isRetVal,
        Direction: direction,
        Type: ty,
        Array: arr,
    }

    cvt.IsArrayType = !cvt.Array.IsZero()
    cvt.IsGList = (cvt.Type.Name == "GLib.List" && !cvt.Type.SubType.IsZero())

    // gotype
    var gotype string
    if cvt.IsArrayType {
        cvt.ElemGoType = getGoType(direction, &cvt.Array.ElemType)
        gotype = "[]" + cvt.ElemGoType
    } else if cvt.IsGList {
        cvt.ElemGoType = "*" + getGoType(direction, &cvt.Type.SubType)
        gotype = "[]" + cvt.ElemGoType
    } else {
        gotype = getGoType(direction, cvt.Type)
    }
    cvt.GoType = gotype

    // cgoType
    var cgoType string
    if cvt.IsArrayType {
        cgoType = getCgoType(direction, &mygi.Type{CType: cvt.Array.CType})
        cvt.ElemCgoType = strings.TrimPrefix(cgoType, "*")
    } else {
        cgoType = getCgoType(direction, cvt.Type)
    }
    cvt.CgoType = cgoType

    return cvt
}

func getConvertExp(valName, tyFrom, tyTo string, cvtMap map[string]string) string {
    var exp string
    key := tyFrom + "," + tyTo
    if method, ok := cvtMap[key]; ok {
        exp = method
    } else {
        // direct cvt
        exp = "(" + tyTo + ")($_)"
    }
    return strings.Replace(exp, "$_", valName, 1)
}

func (cvt *TypeConverter) Go2Cgo(valName string) string {
    return getConvertExp(valName, cvt.GoType, cvt.CgoType, typeConvertMapGo2Cgo)
}

func (cvt *TypeConverter) ElemGo2Cgo(valName string) string {
    return getConvertExp(valName, cvt.ElemGoType, cvt.ElemCgoType, typeConvertMapGo2Cgo)
}

func (cvt *TypeConverter) Cgo2Go(valName string) string {
    return getConvertExp(valName, cvt.CgoType, cvt.GoType, typeConvertMapCgo2Go)
}

func (cvt *TypeConverter) ElemCgo2Go(valName string) string {
    return getConvertExp(valName, cvt.ElemCgoType, cvt.ElemGoType, typeConvertMapCgo2Go)
}

// print convert
// target cgo, source go
func (cvt *TypeConverter) ConvertGo2Cgo(target, source string) {
    if cvt.IsArrayType {
        fmt.Printf("\t// array convert target: %s, source: %s\n", target, source)
        fmt.Printf("\t// array is zero terminated? %v\n", cvt.Array.ZeroTerminated)

        if cvt.Array.ZeroTerminated {
            fmt.Printf("\t%s = (%s)(C.malloc(C.size_t(int(unsafe.Sizeof(*%s)) * (len(%s)+1) )))\n",
                target, cvt.CgoType, target, source)
            fmt.Printf("\tfor i, e := range %s {\n", source)
             elemConvertExp := cvt.ElemGo2Cgo("e")
            fmt.Printf("\t\t%s := %s\n", target + "_E", elemConvertExp)
            fmt.Printf("\t\t*(%s)(unsafe.Pointer(uintptr(unsafe.Pointer(%s)) + uintptr(i)*unsafe.Sizeof(*%s))) = %s\n", cvt.CgoType, target, target, target + "_E")
            fmt.Println("\t}") //end for
            fmt.Printf("\t*(%s)(unsafe.Pointer(uintptr(unsafe.Pointer(%s)) + uintptr(len(%s))*unsafe.Sizeof(*%s))) = nil\n", cvt.CgoType, target, source, target)
        } else {
            fmt.Println("// TODO")
        }
    } else {
        fmt.Printf("\t%s = %s\n", target, cvt.Go2Cgo(source))
    }
}

func (cvt *TypeConverter) ConvertCgo2Go(target, source string) {
    if cvt.IsArrayType {
        fmt.Printf("\t// array convert target: %s, source: %s\n", target, source)

        fmt.Printf("\t// array is zero terminated? %v\n", cvt.Array.ZeroTerminated)

        if cvt.Array.ZeroTerminated {
            varLengthName := source + "_Len"
            fmt.Printf("\t%s := mygibase.ZeroTerminatedArrayLength(unsafe.Pointer(%s))\n", varLengthName, source)
            fmt.Printf("\t%s = make(%s, %s)\n", target, cvt.GoType, varLengthName)
            fmt.Printf("\tfor i := uintptr(0); i < %s; i++ {\n", varLengthName)
            fmt.Printf("\t\t%s := *(%s)(unsafe.Pointer( uintptr(unsafe.Pointer(%s)) + i*unsafe.Sizeof(*%s) ))\n", target + "_E", cvt.CgoType, source, source)
            elemConvertExp := cvt.ElemCgo2Go(target + "_E")
            fmt.Printf("\t\t%s[i] = %s\n", target, elemConvertExp)
            fmt.Println("\t}") // end for
        } else {
            fmt.Println("// TODO")
        }

    } else {
        fmt.Printf("\t%s = %s\n", target, cvt.Cgo2Go(source))
    }
}

func (cvt *TypeConverter) String() string {
    var arrayOrType string
    if cvt.Type.IsZero() {
        arrayOrType = fmt.Sprintf("%#v\n", cvt.Array)
    } else {
        arrayOrType = fmt.Sprintf("%#v\n", cvt.Type)
    }
    return fmt.Sprintf("%#v\n", cvt) +
        //fmt.Sprintf("%#v\n", cvt.Array) +
        //fmt.Sprintf("%#v\n", cvt.Type) +
        arrayOrType +
        fmt.Sprintf("gotype: %s\n", cvt.GoType)
}

func removeCQualifiers(ctype string) string {
    ctype = strings.Replace(ctype, "const ", "", 1)
    ctype = strings.Replace(ctype, "volatile ", "", 1)
    ctype = strings.Replace(ctype, " const", "", 1)
    ctype = strings.Replace(ctype, " volatile", "", 1)
    return ctype
}

func getGoType(direction string, ty mygi.TypeLike) string {
    name := ty.GetName()
    ctype := removeCQualifiers(ty.GetCType())

    if direction == "out" {
        ctype = strings.TrimSuffix(ctype, "*")
    }
    //fmt.Printf("getGoType name %q, ctype %q\n", name, ctype)
    if name == "utf8" || name == "filename" {
        return "string"
    }
    if name == "GError" {
        return "error"
    }
    nStar := cTypeStarCount( ctype )
    ctype = removeCTypeStar(ctype)
    return repeatStar(nStar) + _getGoType(name, ctype)
}

func _getGoType(name, ctype string) string {
    var gotype string
    if ctype != "" {
        gotype = baseType2GoType(ctype)
    } else {
        gotype = baseType2GoType(name)
    }

    if gotype != "" {
        return gotype
    }
    // ctype GVariant
    // name GLib.Variant
    return getGoTypeByTypeName(name)
}

func getCgoType(direction string, ty mygi.TypeLike) string {
    //name := ty.GetName()
    ctype := removeCQualifiers(ty.GetCType())
    if direction == "out" {
        ctype = strings.TrimSuffix(ctype, "*")
    }
    //fmt.Printf("getCgoType ctype %q\n", ctype)
    if ctype == "void*" {
        return "unsafe.Pointer"
    }

    nStar := cTypeStarCount(ctype)
    ctypeWithoutStar := removeCTypeStar(ctype)

    return repeatStar(nStar) + "C." + ctypeWithoutStar
}



// "string,char*" C.Cstring
// "string,gchar*" mygibase.GoString2GString
// "boolean,gboolean" C.gboolean(mygibase.Bool2Int(x))
// 直接转换 利用 C 的 CGO 的表示方法
// "mygibase.Gint,gint" -> C.gint
// go 先获取unsafe pointer之后直接转换
// "*File,gconstpointer" C.gconstpointer(this.GetUnsafePointer()
// "*File,GFile*" (*C.GFile)(this.GetUnsafePointer())

// key is "go,cgo", value is method
var typeConvertMapGo2Cgo map[string]string
// key is "cgo,go", value is method
var typeConvertMapCgo2Go map[string]string

func init() {
    // C_ret -> Go_ret
    typeConvertMapCgo2Go = map[string]string{
        "C.gboolean,bool": "$_ != 0",
        "*C.char,string":"C.GoString($_)",
        "*C.gchar,string": "C.GoString((*C.char)(unsafe.Pointer($_)))",
        "*C.GError,error": "mygibase.GError2Error(unsafe.Pointer($_))",
    }
    // Go_ret := bool(C_ret != 0)
    // Go_ret := C.GoString(C_ret)
    // Go_ret := C.GoString((*C.char)(unsafe.Pointer(C_ret)))

    // 直接转换
    // Go_ret := float32(C_ret) C.gfloat
    // Go_ret := float64(C_ret) C.gdouble
    // Go_ret := mygibase.Guint(C_ret)

    // arg -> C_arg for c call
    typeConvertMapGo2Cgo = map[string]string{
        "bool,C.gboolean": "C.gboolean(mygibase.Bool2Num($_))",
        "string,*C.char": "C.CString($_)",
        "string,*C.gchar": "(*C.gchar)(unsafe.Pointer(C.CString($_)))",
    }
    // C_arg := C.gboolean(mygibase.Bool2Num(arg))
    // C_arg := C.CString(arg)
    // C_arg := (*C.gchar)(C.CString(arg))

    // ctype GFile*
    // C_arg := (*C.GFile)(arg.GetUnsafePointer())
}

func RegisterConvertCgo2Go(from, to, method string) {
    key := from + "," + to
    fmt.Printf("// RegisterConvertCgo2Go %s %s\n", key, method)
    typeConvertMapCgo2Go[key] = method
}

func RegisterConvertGo2Cgo(from, to, method string) {
    key := from + "," + to
    fmt.Printf("// RegisterConvertGo2Cgo %s %s\n", key, method)
    typeConvertMapGo2Cgo[key] = method
}



func cTypeStarCount(ctype string) int {
    ctypeRunes := []rune(ctype)
    starCount := 0
    for i := len(ctypeRunes) - 1; i >= 0; i-- {
        char := ctypeRunes[i]
        if char != '*' {
            break
        }
        starCount++
    }
    return starCount
}

func removeCTypeStar(ctype string) string {
    return strings.TrimRight(ctype, "*")
}

func repeatStar(n int) string {
    return strings.Repeat("*", n)
}

func baseType2GoType(ctype string) string {
    // remove prefix 'const' then trim space
    ctype = strings.TrimSpace(strings.TrimPrefix(ctype, "const"))
    switch ctype {
        case "gboolean":
            return "bool"
        case "gpointer", "gconstpointer":
            return "unsafe.Pointer"
        case "gfloat", "float":
            return "float32"
        case "gdouble", "double":
            return "float64"
        case "GType":
            return "mygibase.GType"
        case "gchar", "guchar", "char",
            "gunichar", "gunichar2",
            "gshort", "gushort",
            "gint", "guint",
            "glong", "gulong",
            "goffset", "gsize", "gssize",
            "gintptr", "guintptr",
            "time_t":
            return "mygibase." + snake2Camel(ctype)
        case "int":
            return "mygibase.Gint"
        case "gint8", "guint8",
            "gint16", "guint16",
            "gint32", "guint32",
            "gint64", "guint64":
            return ctype[1:]
        default:
            // fail
            return ""
    }
}


// get gotype by Type.Name
func getGoTypeByTypeName(name string) string {
    gotype := name
    if strings.ContainsRune(name, '.') {
        nameArr := strings.SplitN(name, ".", 2)
        ns := strings.ToLower(nameArr[0])
        if ns == namespace {
            // GLib.HashTable -> HashTable if namespace == "glib"
            gotype = nameArr[1]
        } else {
            gotype = ns + "." + nameArr[1]
            // GLib.Variant -> glib.Variant
        }
    }
    return gotype
}

// snake_case to CamelCase
func snake2Camel(name string) string {
    //name = strings.ToLower(name)
	var out bytes.Buffer
	for _, word := range strings.Split(name, "_") {
		word = strings.ToLower(word)
		//if subst, ok := config.word_subst[word]; ok {
			//out.WriteString(subst)
			//continue
		//}

		if word == "" {
			out.WriteString("_")
			continue
		}
		out.WriteString(strings.ToUpper(word[0:1]))
		out.WriteString(word[1:])
	}
	return out.String()
}
