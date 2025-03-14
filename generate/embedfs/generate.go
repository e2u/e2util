// generate.go
// example:
// go run github.com/e2u/e2util/generate/embedfs -source ./assets -xxx
// or in main.go
// //go:generate go run github.com/e2u/e2util/generate/embedfs -source ./assets
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Options struct {
	FilePatterns []string
	Output       string
	Package      string
	Directory    string
}

const sourceTemplate = `// Code generated by github.com/e2u/e2util/generate/embedfs/generate.go; DO NOT EDIT.

package {{.Package}}

import (
	"embed"
)

{{range .Directories}}
//go:embed {{.}}{{end}}
var {{.VarName}} embed.FS
`

func main() {
	filePatterns := flag.String("patterns", "*.js,*.css", "Comma-separated list of file patterns")
	output := flag.String("output", "generated.go", "Output file name")
	pkg := flag.String("package", "assets", "Package name")
	source := flag.String("source", ".", "Source Directory to embed files from")
	varName := flag.String("name", "EmbedFS", "Name of the file to embed")

	flag.Parse()

	patterns := splitPatterns(*filePatterns)

	sourceDeep := len(strings.Split(*source, string(filepath.Separator)))

	dirsMap := make(map[string]struct{})

	for _, pattern := range patterns {
		files, err := collectFiles(pattern, *source)
		if err != nil {
			fmt.Println("Error collecting files:", err)
			return
		}

		for _, f := range files {
			dir := filepath.Dir(f)

			var ts []string
			for range strings.Split(dir, string(filepath.Separator)) {
				ts = append(ts, "**")
			}
			ts = append(ts, pattern)
			if len(ts) > 0 {
				ts = ts[sourceDeep-1:]
				dirsMap[strings.Join(ts, string(filepath.Separator))] = struct{}{}
			}
		}
	}

	var dirsArray []string
	for d := range dirsMap {
		dirsArray = append(dirsArray, d)
	}

	f, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tmpl, err := template.New("source").Parse(sourceTemplate)
	if err != nil {
		panic(err)
	}

	if err := tmpl.Execute(f, map[string]interface{}{
		"Package":     *pkg,
		"VarName":     *varName,
		"Directories": dirsArray,
	}); err != nil {
		panic(err)
	}

	fmt.Printf("Generated %s\n", *output)
}

// collectFiles 根據文件模式收集文件
func collectFiles(pattern string, dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return err
		}
		if matched && !info.IsDir() {
			files = append(files, path)
			return nil
		}
		//}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// splitPatterns 將逗號分隔的字串轉換為切片
func splitPatterns(patterns string) []string {
	var result []string
	for _, p := range strings.Split(patterns, ",") {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}
