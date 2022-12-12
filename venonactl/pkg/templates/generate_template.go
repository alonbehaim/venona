/*
Copyright 2019 The Codefresh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

/*
for usage in
\\go:generate go run generate generate_template.go <folder name under templates>
reads all files in folder and appends them to template map
*/

var outfileBaseName = "templates.go"
var packageTemplate = template.Must(template.New("").Parse(
	`
// Code generated by go generate; DO NOT EDIT.
// using data from templates/{{ .FolderName }}
package {{ .PackageName }}

func TemplatesMap() map[string]string {
    templatesMap := make(map[string]string)` +

		"\n{{ range $key, $value := .TemplateFilesMap }}" +
		"\ntemplatesMap[\"{{ $key }}\"] = `{{ $value }}` \n" +
		"{{ end }}" + `
    return  templatesMap
}
`))

type packageTempateData struct {
	PackageName      string
	FolderName       string
	TemplateFilesMap map[string]string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("generate_template ERROR: missing folder name")
		os.Exit(1)
	}

	var currentFilePath string
	if strings.Contains(os.Args[0], "/go-build") {
		_, currentFilePath, _, _ = runtime.Caller(0)
	} else {
		currentFilePath = os.Args[0]
	}

	currentDir := filepath.Dir(currentFilePath)
	templatesDirParam := os.Args[1]
	var folderName = path.Join(currentDir, templatesDirParam)

	// Fill Tempalte Map
	templateFilesMap := make(map[string]string)
	filepath.Walk(folderName, func(name string, info os.FileInfo, err error) error {
		if !info.IsDir() && path.Base(name) != outfileBaseName {
			b, _ := os.ReadFile(name)
			templateFilesMap[filepath.Base(name)] = string(b)
		}
		return nil
	})

	if len(templateFilesMap) == 0 {
		fmt.Printf("No files in %s\n", folderName)
	}

	outfileName := path.Join(folderName, "templates.go")
	outfile, err := os.Create(outfileName)
	if err != nil {
		fmt.Printf("generate_template ERROR: cannot create out file %s, %v \n", outfileName, err)
		os.Exit(1)
	}
	defer outfile.Close()

	err = packageTemplate.Execute(outfile, packageTempateData{
		PackageName:      path.Base(folderName),
		FolderName:       templatesDirParam,
		TemplateFilesMap: templateFilesMap,
	})
	if err != nil {
		fmt.Printf("generate_template ERROR: cannot generate template %v \n", err)
		os.Exit(1)
	}
}
