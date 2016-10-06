package main

import (
    "fmt"
    "os"
    "mygi"
    "strings"
    "io/ioutil"
    "path/filepath"
    "gopkg.in/yaml.v2"
    "./typeconverter"
)

type LibConfig struct {
    Namespace string `yaml:"Namespace"`
    Version string `yaml:"Version"`
    CIncludes []string `yaml:"CIncludes"`
}

var namespace string

func main() {
    libDir := os.Args[1]
    // read lib.in config.yml
    var libCfg LibConfig
    libCfgBytes, err := ioutil.ReadFile(filepath.Join(libDir, "config.yml"))
    if err != nil {
        fmt.Println("err:", err)
        return
    }

    err = yaml.Unmarshal(libCfgBytes, &libCfg)
    if err != nil {
        fmt.Println("err:", err)
        return
    }
    namespace = libCfg.Namespace
    version := libCfg.Version

    // package
    pkg := strings.ToLower(namespace)
    fmt.Println("package", pkg)

    fmt.Println("// namespace:", namespace)
    fmt.Println("// version:", version)

    // parse gir xml
    repo, err := mygi.Load(namespace, version)
    if err != nil {
        panic(err)
    }
    fmt.Println("// parse gir xml ok")

    // c includes
    fmt.Println("\n/*")
    for _, ci := range repo.CIncludes() {
        fmt.Printf("#include <%s>\n", ci.Name)
    }
    fmt.Printf("#include <%s>\n", "stdlib.h")
    for _, ci := range libCfg.CIncludes {
        fmt.Printf("#include <%s>\n", ci)
    }
    // pkg-config
    var pkgs []string
    for _, p := range repo.Packages {
        pkgs = append(pkgs, p.Name)
    }
    fmt.Printf("#cgo pkg-config: %s\n", strings.Join(pkgs, " "))
    fmt.Println("*/\nimport \"C\"")

    // go import
    fmt.Println(
`import (
    "unsafe"
    mygibase "mygi/base"
)

var _ unsafe.Pointer
var _ mygibase.Gchar
`)

    repons := repo.Namespace

    // alias
    for _, alias := range repons.Aliases {
        pAlias(&alias)
    }

    // enum
    for _, enum := range repons.Enums {
        pEnum(&enum)
    }

    // bitfield
    for _, enum := range repons.Bitfields {
        pEnum(&enum)
    }

    // constant
    fmt.Println("\n// constants")
    for _, constant := range repons.Constants {
        pConstant(&constant)
    }

    // struct
    for _, st := range repons.Records {
        pStruct(&st)
    }

    // function
    for _, f := range repons.Functions {
        pFunction(&f)
    }

    // registerConvert
    for _, c := range repons.Classes {
        registerConvert(c.Name, c.CType)
    }
    for _, ifc := range repons.Interfaces {
        registerConvert(ifc.Name, ifc.CType)
    }


    // class
    for _, c := range repons.Classes {
        pClass(&c)
    }

    // interface
    for _, ifc := range repons.Interfaces {
        pInterface(&ifc)
    }
}

func pAlias(alias *mygi.Alias) {
    fmt.Println("\n// alias")
    fmt.Printf("type %s C.%s\n", alias.Name, alias.CType)
}

func pEnum(enum *mygi.Enum) {
    fmt.Println("\n// enum")
    // type def
    fmt.Printf("type %s C.%s\n", enum.Name, enum.CType)
    fmt.Println("const (")
    for i, member := range enum.Members {
        goValName := enum.Name + snake2Camel(member.Name)
        if i == 0 {
            fmt.Printf("\t%s %s = %s\n", goValName, enum.Name, member.Value)
        } else {
            fmt.Printf("\t%s = %s\n", goValName, member.Value)
        }
    }
    fmt.Println(")")
}

func pConstant(c *mygi.Constant) {
    // NOTE: no use c.CType c.CType == C.Name
    //fmt.Println("// const ctype:", c.CType)
    //fmt.Printf("// constant type %#v\n", c.Type)
    var value string
    switch c.Type.Name {
    case "utf8":
        // quote string
        value = fmt.Sprintf("%#v", c.Value)
    case "gint":
        value = c.Value
    default:
        panic("unsupport constant type")
    }
    fmt.Printf("const %s = %s\n", snake2Camel(strings.ToLower(c.Name)), value)
}

func pStruct(s *mygi.Record) {
    if s.GlibIsGtypeStructFor != "" {
        fmt.Println("// struct ignore", s.Name)
        return
    }
    if s.Disguised {
        fmt.Println("// struct ignore", s.Name)
        return
    }

    fmt.Println("// struct")
    fmt.Printf("type %s %s\n", s.Name, "C." + s.CType)

    // struct methods
    for _, method := range s.Methods {
        pMethod(&method)
    }
}

func registerConvert(typeName, cType string) {
    cgoType := "*C." + cType
    goType := "*" + typeName
    // *C.GAppInfo,*AppInfo
    typeconverter.RegisterConvertCgo2Go(cgoType, goType,
        fmt.Sprintf("%sNewWithUnsafePointer(unsafe.Pointer($_))", typeName))

    // *AppInfo,*C.GAppInfo
    typeconverter.RegisterConvertGo2Cgo(goType, cgoType,
        fmt.Sprintf("(%s)($_.GetUnsafePointer())", cgoType) )
}

func pClass(c *mygi.Class) {
    fmt.Println("\n// class")

    // type struct block
    fmt.Printf("type %s struct {\n", c.Name)
    fmt.Println("\tC unsafe.Pointer")
    // ifcs
    var ifcs []string
    for _, ifc := range c.ImplementedInterfaces {
        ifcs = append(ifcs, ifc.Name)
        fmt.Printf("\t%s\n", ifc.Name)
    }
    fmt.Println("}")

    // GetUnsafePointer method
    pGetUnsafePointerMethod(c.Name)

    //  c.Name + NewWithUnsafePointer func
    pNewWithUnsafePointerFunc(c.Name, ifcs)

    // class constructors
    for _, constructor := range c.Constructors {
        pConstructor(&constructor, c.Name)
    }

    // class methods
    for _, method := range c.Methods {
        pMethod(&method)
    }
}
//func (this *AppInfo) GetUnsafePointer() unsafe.Pointer {
    //return this.C
//}
func pGetUnsafePointerMethod(name string) {
    fmt.Printf("\nfunc (this *%s) GetUnsafePointer() unsafe.Pointer {\n", name)
    fmt.Printf("\treturn this.C\n}\n")
}

//func DesktopAppInfoNewWitchUnsafePointer(p unsafe.Pointer) *AppInfo {
    //this := &AppInfo{C:p}
    //this.AppInfo.C = p
    //return this
//}
func pNewWithUnsafePointerFunc(name string, ifcs []string) {
    fmt.Printf("\nfunc %sNewWithUnsafePointer(p unsafe.Pointer) *%s {\n", name, name)
    fmt.Printf("\tthis := &%s{C:p}\n", name)
    // set ifc.C
    for _, ifc := range ifcs {
        fmt.Printf("\tthis.%s.C = p\n", ifc)
    }
    fmt.Println("\treturn this\n}")
}

func pInterface(ifc *mygi.Interface) {
    fmt.Println("\n// ifc")

    fmt.Printf("type %s struct {\n", ifc.Name)
    fmt.Println("\tC unsafe.Pointer")
    fmt.Println("}")

    pGetUnsafePointerMethod(ifc.Name)
    pNewWithUnsafePointerFunc(ifc.Name, nil)

    // interface methods
    for _, method := range ifc.Methods {
        pMethod(&method)
    }
}

