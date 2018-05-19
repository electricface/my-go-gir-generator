package main

import (
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/electricface/my-go-gir-generator/gi"
	"log"
)

func main() {
	flag.Parse()
	namespace := flag.Arg(0)
	version := flag.Arg(1)
	repo, err := gi.Load(namespace, version)
	if err != nil {
		log.Fatal(err)
	}
	_ = repo

	funcName := flag.Arg(2)
	fmt.Println("funcName is:", funcName)

	ns := repo.Namespace
	funcInfo := ns.GetFunctionInfo(funcName)
	if funcInfo == nil {
		log.Fatal("not found func ", funcName)
	}

	// print func info detail
	fmt.Println("found", funcInfo.CIdentifier)

	if funcInfo.Deprecated {
		fmt.Println("deprecated since version:", funcInfo.DeprecatedVersion)
	}

	if funcInfo.Throws {
		fmt.Println("throws error")
	}

	//funcInfo.Parameters.InstanceParameter
	//funcInfo.Parameters.Parameters
	fmt.Println("Return Value:")
	spew.Dump(funcInfo.ReturnValue)

	if funcInfo.Parameters != nil {

		if funcInfo.Parameters.InstanceParameter != nil {
			fmt.Println("\nInstance Parameter:")
			spew.Dump(funcInfo.Parameters.InstanceParameter)
		}

		if len(funcInfo.Parameters.Parameters) > 0 {
			fmt.Println("\nParameters:")
		}
		for _, param := range funcInfo.Parameters.Parameters {
			spew.Dump(param)
		}
	}
}
