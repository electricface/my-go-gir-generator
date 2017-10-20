package main

import (
	"mygi"
	"log"
	"os"
)

var libCfg *LibConfig

func main() {
	repo, err := mygi.Load("Gio", "2.0")
	if err != nil {
		log.Fatal(err)
	}

	types := repo.GetTypes()
	log.Print(len(types))

	for name, type0 := range types {
		log.Printf("%s -> %T\n", name, type0)
	}

	os.Exit(0)
	interfaces := repo.Namespace.Interfaces
	for _, interface0 := range interfaces {
		if interface0.Name == "AppInfo" {
			sourceFile := NewSourceFile("gio")
			pInterface(sourceFile, interface0)
			//sourceFile.Print()
			sourceFile.Save("out/appinfo.go")
		}
	}
}

func pInterface(s *SourceFile, interface0 *mygi.Interface) {
	name := interface0.Name
	s.GoBody.Pn("// interface %s", name)

	s.GoBody.Pn("type %s struct {", name )
	s.GoBody.Pn("Ptr unsafe.Pointer")
	s.GoBody.Pn("}")

	cPtrType := "*C." + interface0.CType

	// method native
	s.GoBody.Pn("func (v %s) native() %s {", name, cPtrType)
	s.GoBody.Pn("return (%s)(v.Ptr)", cPtrType)
	s.GoBody.Pn("}")

	// method wrapXXX
	s.GoBody.Pn("func wrap%s(p %s) %s {", name, cPtrType, name )
	s.GoBody.Pn("return %s{unsafe.Pointer(p)}", name)
	s.GoBody.Pn("}")

	// method WrapXXX
	s.GoBody.Pn("func Wrap%s(p unsafe.Pointer) %s {", name, name )
	s.GoBody.Pn("return %s{p}", name)
	s.GoBody.Pn("}")

	// methods
	for _, method := range interface0.Methods {
		s.GoBody.Pn("// method %s", method.Name())
		s.GoBody.Pn("// %s", method.CIdentifier)

		if method.CIdentifier == "g_app_info_get_id" {
			//pMethod(&method)
		}

	}

}