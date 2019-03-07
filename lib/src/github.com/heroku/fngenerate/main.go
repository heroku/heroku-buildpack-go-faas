package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

var (
	ErrMissingName    = errors.New("Missing `name`")
	ErrMissingPath    = errors.New("Missing `path`")
	ErrInvalidTrigger = errors.New("Invalid `trigger` value")
)

type YAML struct {
	Build struct {
		Functions map[string]map[string]string `functions`
	} `yaml:"build"`
}

type Params struct {
	ImportPath string
	HTTP       []Route
}

type Route struct {
	Name string
	Path string
}

func main() {
	importPath := flag.String("i", "", "Customer import path")
	yamlPath := flag.String("y", "", "Customer heroku.yaml path")
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

	file, err := os.Open(*yamlPath)
	FatalIf(err)

	var yml YAML
	yaml.NewDecoder(file).Decode(&yml)
	FatalIf(err)

	params := Params{
		ImportPath: *importPath,
	}

	for _, f := range yml.Build.Functions {
		switch f["trigger"] {
		case "http":
			if f["name"] == "" {
				FatalIf(ErrMissingName)
			}

			if f["path"] == "" {
				FatalIf(ErrMissingPath)
			}

			path := f["path"]
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			params.HTTP = append(params.HTTP, Route{
				Name: f["name"],
				Path: path,
			})

			break
		default:
			FatalIf(ErrInvalidTrigger)
		}
	}

	tmpl.Execute(os.Stdout, params)
}

func FatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
