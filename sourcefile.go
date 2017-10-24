package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"mygi"
)

type SourceFile struct {
	//Filename string
	Pkg       string
	CPkgs     []string
	CIncludes []string
	CBody     *SourceBody

	GoImports []string
	GoBody    *SourceBody
}

func NewSourceFile(pkg string) *SourceFile {
	return &SourceFile{
		Pkg: pkg,

		CBody:  &SourceBody{},
		GoBody: &SourceBody{},
	}
}

func (v *SourceFile) Print() {
	v.WriteTo(os.Stdout)
}

func (v *SourceFile) Save(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	v.WriteTo(f)

	err = f.Sync()
	if err != nil {
		panic(err)
	}
}

func (v *SourceFile) WriteTo(w io.Writer) {
	io.WriteString(w, "package "+v.Pkg+"\n")

	if len(v.CPkgs) > 0 ||
		len(v.CIncludes) > 0 ||
		len(v.CBody.buf.Bytes()) > 0 {

		io.WriteString(w, "/*\n")
		if len(v.CPkgs) != 0 {
			str := "#cgo pkg-config: " + strings.Join(v.CPkgs, " ") + "\n"
			io.WriteString(w, str)
		}

		for _, inc := range v.CIncludes {
			io.WriteString(w, "#include "+inc+"\n")
		}

		w.Write(v.CBody.buf.Bytes())

		io.WriteString(w, "*/\n")
		io.WriteString(w, "import \"C\"\n")
	}

	for _, imp := range v.GoImports {
		io.WriteString(w, "import "+imp+"\n")
	}

	w.Write(v.GoBody.buf.Bytes())
}

// unsafe => "unsafe"
// or x,github.com/path/ => x "path"
func (s *SourceFile) AddGoImport(imp string) {
	log.Println("SourceFile.AddGoImport:", imp)
	var importStr string
	if strings.Contains(imp, ",") {
		parts := strings.SplitN(imp, ",", 2)
		importStr = fmt.Sprintf("%s %q", parts[0], parts[1])
	} else {
		importStr = `"` + imp + `"`
	}

	for _, imp0 := range s.GoImports {
		if imp0 == importStr {
			return
		}
	}
	s.GoImports = append(s.GoImports, importStr)
}

func (s *SourceFile) AddGirImport(ns string) {
	repo := mygi.GetLoadedRepo(ns)
	if repo == nil {
		panic("failed to get loaded repo " + ns)
	}
	base := strings.ToLower(repo.Namespace.Name) + "-" + repo.Namespace.Version
	fullPath := "github.com/electricface/go-auto-gir/" + base
	s.AddGoImport(fullPath)
}

type SourceBody struct {
	buf bytes.Buffer
}

func (v *SourceBody) Pn(format string, a ...interface{}) {
	str := fmt.Sprintf(format, a...)
	v.buf.WriteString(str)
	v.buf.WriteByte('\n')
}

func (v *SourceBody) P(format string, a ...interface{}) {
	str := fmt.Sprintf(format, a...)
	v.buf.WriteString(str)
}
