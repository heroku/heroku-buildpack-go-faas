package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	toml "github.com/pelletier/go-toml"
	"github.com/pelletier/go-toml/query"
)

var (
	ErrMissingFunc = errors.New("Missing `function`")
	ErrMissingPath = errors.New("Missing `path`")
	ErrInvalidFunc = errors.New("Invalid `function` value")
	ErrInvalidPath = errors.New("Invalid `path` value")
)

const FuncQuery = "$.metadata.heroku.functions"

type Params struct {
	ImportPath string
	Routes     []Route
}

type Route struct {
	Function string
	Path     string
}

func main() {
	importPath := flag.String("i", "", "Customer import path")
	tomlPath := flag.String("g", "", "Customer Gopkg.toml path")
	templatePath := flag.String("t", "", "Template file path")
	flag.Parse()

	if len(os.Args) == 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	content, err := ioutil.ReadFile(*templatePath)
	FatalIf(err)

	tmpl, err := template.New("main").Parse(string(content))
	FatalIf(err)

	tree, err := toml.LoadFile(*tomlPath)
	FatalIf(err)

	results, err := query.CompileAndExecute(FuncQuery, tree)
	FatalIf(err)

	params := Params{
		ImportPath: *importPath,
		Routes:     []Route{},
	}

	for _, v := range results.Values() {
		trees := v.([]*toml.Tree)
		for _, t := range trees {
			function, ok := t.Get("function").(string)
			if !ok {
				FatalIf(ErrInvalidFunc)
			}

			if function == "" {
				FatalIf(ErrMissingFunc)
			}

			path, ok := t.Get("path").(string)
			if !ok {
				FatalIf(ErrInvalidPath)
			}

			if path == "" {
				FatalIf(ErrMissingPath)
			}

			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			params.Routes = append(params.Routes, Route{
				Function: function,
				Path:     path,
			})
		}
	}

	tmpl.Execute(os.Stdout, params)
}

func FatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
