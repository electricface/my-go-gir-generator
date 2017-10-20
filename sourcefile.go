package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type SourceFile struct {
	//Filename string
	Pkg string
	CPkgs []string
	CIncludes []string
	CBody *SourceBody

	GoImports []string
	GoBody *SourceBody
}

func NewSourceFile(pkg string) *SourceFile {
	return &SourceFile{
		Pkg: pkg,

		CBody: &SourceBody{},
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
	io.WriteString(w, "package " + v.Pkg + "\n")

	if  len(v.CPkgs) > 0 ||
		len(v.CIncludes) > 0 ||
			len(v.CBody.buf.Bytes()) > 0 {

		io.WriteString(w, "/*\n")
		if len(v.CPkgs) != 0 {
			str := "#cgo pkg-config: " + strings.Join(v.CPkgs, " ") + "\n"
			io.WriteString(w, str)
		}

		for _, inc := range v.CIncludes{
			io.WriteString(w, "#include " + inc + "\n")
		}

		w.Write(v.CBody.buf.Bytes())

		io.WriteString(w, "*/\n")
		io.WriteString(w, "import \"C\"\n")
	}

	for _, imp := range v.GoImports {
		io.WriteString(w, "import " + imp + "\n")
	}

	w.Write(v.GoBody.buf.Bytes())
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

