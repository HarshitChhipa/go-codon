package golang

import (
	"os"
	"fmt"
	"log"
	"errors"
	"strings"
	"path/filepath"
	goruntime "runtime"

	swagger_generator "github.com/go-swagger/go-swagger/generator"
)

func GenerateUpstreamSwagger(gen *generator) bool {
	gen.CurrentAPIName = strings.TrimSuffix(gen.CurrentSpecFile, ".yml")
	gen.CurrentAPIName = strings.TrimSuffix(gen.CurrentAPIName, ".yaml")

	gen.CurrentDirTarget = fmt.Sprintf("clients/%s/", gen.CurrentAPIName)

	opts := &swagger_generator.GenOpts{
		APIPackage:        "operations",
		ModelPackage:      "models",
		ServerPackage:     "restapi",
		ClientPackage:     "client",
		Principal:         "",
		DefaultScheme:     "http",
		DefaultProduces:   "application/json",
		IncludeModel:      true,
		IncludeValidator:  true,
		IncludeHandler:    true,
		IncludeParameters: true,
		IncludeResponses:  true,
		ValidateSpec:      true,
		Tags:              []string{},
		IncludeSupport:    true,
		DumpData:          false,
		Spec:              gen.CurrentSpecFilePath,
		Target:            gen.CurrentDirTarget,
		TemplateDir:       "spec/templates/",
	}
	if err := opts.EnsureDefaults(true); err != nil {
		log.Println(err)
		return false
	}
	if err := swagger_generator.GenerateClient("", []string{}, []string{}, opts); err != nil {
		log.Println(err)
		return false
	}

	import_path, err := baseImport(filepath.Join(gen.CurrentDirTarget, opts.ClientPackage))
	if err != nil {
		log.Println(err)
		return false
	}
	gen.ClientImports[gen.CurrentAPIName] = import_path

	return true
}

func GenerateServiceSwagger(gen *generator) bool {
	// swagger generate server -f spec/server/main.yml -t server -T spec/templates/
	gen.CurrentDirTarget = "server"
	gen.CurrentSpecFilePath = "spec/server/main.yml"

	opts := &swagger_generator.GenOpts{
		IncludeModel:      true,
		IncludeValidator:  true,
		IncludeHandler:    true,
		IncludeParameters: true,
		IncludeResponses:  true,
		IncludeURLBuilder: true,
		IncludeMain:       true,
		IncludeSupport:    true,
		ExcludeSpec:       false,
		DumpData:          false,
		WithContext:       false,
		ValidateSpec:      true,
		Spec:              gen.CurrentSpecFilePath,
		APIPackage:        "operations",
		ModelPackage:      "models",
		ServerPackage:     "restapi",
		ClientPackage:     "client",
		Principal:         "",
		Target:            gen.CurrentDirTarget,
		DefaultScheme:     "http",
		DefaultProduces:   "",
		Tags:              []string{},
		TemplateDir:       "spec/templates/",
		Models:            []string{},
		Operations:        []string{},
		Name:              "",
		FlagStrategy:      "go-flags",
		CompatibilityMode: "modern",
	}
	if err := opts.EnsureDefaults(false); err != nil {
		log.Println(err)
		return false
	}

	// TODO: Add support for template config
	// Adding views template to configuration
	opts.Sections.Application = append(opts.Sections.Application, swagger_generator.TemplateOpts{
		Name:       "views",
		Source:     "asset:serverViews",
		Target:     "{{ joinFilePath .Target .ServerPackage }}",
		FileName:   "views.go",
	})

	if err := swagger_generator.GenerateServer("", []string{}, []string{}, opts); err != nil {
		log.Println(err)
		return false
	}

	return true
}

// Copyright 2015 go-swagger maintainers
// Use of this source code is governed by Apache License,
// Version 2.0 that can be found in the LICENSE file.
// Modified error reporting structure to match go-codon's
func baseImport(tgt string) (string, error) {
	p, err := filepath.Abs(tgt)
	if err != nil {
		return "", err
	}

	var pth string
	for _, gp := range filepath.SplitList(os.Getenv("GOPATH")) {
		pp := filepath.Join(filepath.Clean(gp), "src")
		var np, npp string
		if goruntime.GOOS == "windows" {
			np = strings.ToLower(p)
			npp = strings.ToLower(pp)
		}
		if strings.HasPrefix(np, npp) {
			pth, err = filepath.Rel(pp, p)
			if err != nil {
				return "", err
			}
			break
		}
	}

	if pth == "" {
		return "", errors.New("target must reside inside a location in the $GOPATH/src")
	}
	return pth, nil
}
