package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
)

const controlFormat = `Source: golang-github-linuxdeepin-go-gir
Section: devel
Priority: extra
Maintainer: Deepin Packages Builder <packages@deepin.com>
Build-Depends: debhelper (>= 10), dh-golang, golang-any,
%s
Standards-Version: 4.0.0
Homepage: https://github.com/linuxdeepin/go-gir
XS-Go-Import-Path: github.com/linuxdeepin/go-gir
Testsuite: autopkgtest-pkg-go

Package: golang-github-linuxdeepin-go-gir-dev
Architecture: all
Depends: ${shlibs:Depends}, ${misc:Depends}
Description: Go bindings based on GObject-Introspection

`

var (
	projectRoot string
	debianDir   string
	configFile  string
)

func init() {
	flag.StringVar(&projectRoot, "project-root", "", "")
	flag.StringVar(&configFile, "cfg", "", "")
}

type Config struct {
	Packages []*Package `toml:"packages"`
}

type Package struct {
	Name    string   `toml:"name"`
	Desc    string   `toml:"desc"`
	Dirs    []string `toml:"dirs"`
	Depends []string `toml:"depends"`
}

func loadCfg(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = toml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	// fix dirs
	for _, pkg := range config.Packages {
		if len(pkg.Dirs) == 0 {
			pkg.Dirs = []string{pkg.Name}
		}
	}

	return &config, nil
}

func main() {
	flag.Parse()
	log.Println("project root:", projectRoot)
	debianDir = filepath.Join(projectRoot, "debian")

	cfg, err := loadCfg(configFile)
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range cfg.Packages {
		writeInstall(pkg)
	}
	writeControl(cfg.Packages)
}

/*
example

Package: golang-github-linuxdeepin-go-gir-gudev-1.0-dev
Architecture: all
Depends: ${shlibs:Depends}, ${misc:Depends}, golang-github-linuxdeepin-go-gir-dev, libgudev-1.0-dev
Description: go bindings for the GUdev library

*/

func writePackage(cfg *Package, buf *bytes.Buffer) {
	pkg := getPackageName(cfg.Name)
	buf.WriteString(fmt.Sprintf("Package: %s\n", pkg))
	buf.WriteString("Architecture: all\n")
	deps := append([]string{"golang-github-linuxdeepin-go-gir-dev"}, cfg.Depends...)
	buf.WriteString(fmt.Sprintf("Depends: ${shlibs:Depends}, ${misc:Depends}, %s\n",
		strings.Join(deps, ", ")))
	buf.WriteString(fmt.Sprintf("Description: go bindings for the %s library\n\n", cfg.Desc))
}

func getPackageName(name string) string {
	return fmt.Sprintf("golang-github-linuxdeepin-go-gir-%s-dev", name)
}

func writeControl(packages []*Package) {
	filename := filepath.Join(debianDir, "control")
	var buf bytes.Buffer

	var deps []string
	for _, pkg := range packages {
		for _, dep := range pkg.Depends {
			deps = append(deps, " "+dep)
		}
	}
	depends := strings.Join(deps, ",\n")
	buf.WriteString(fmt.Sprintf(controlFormat, depends))

	for _, pkg := range packages {
		writePackage(pkg, &buf)
	}

	err := ioutil.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func writeInstall(cfg *Package) {
	name := cfg.Name
	fileBasename := getPackageName(name) + ".install"
	filename := filepath.Join(debianDir, fileBasename)
	var buf bytes.Buffer
	for _, dir := range cfg.Dirs {
		buf.WriteString("usr/share/gocode/src/github.com/linuxdeepin/go-gir/" + dir + "\n")
	}
	err := ioutil.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
