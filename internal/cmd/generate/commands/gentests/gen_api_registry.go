// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

var (
	pkgNames []string
	debug    bool
)

func init() {
	if pkgNamesEnv := os.Getenv("PACKAGE_NAMES"); pkgNamesEnv != "" {
		pkgNames = strings.Split(pkgNamesEnv, ",")
	} else {
		pkgNames = []string{
			"github.com/elastic/go-elasticsearch/v8/esapi",
			"github.com/elastic/go-elasticsearch/v8/esapi/xpack",
		}
	}

	if _, ok := os.LookupEnv(""); ok {
		debug = true
	}
}

func main() {
	log.SetFlags(0)

	log.Println("Generating API registry into api_registry.gen.go")
	log.Println(strings.Repeat("=", 80))

	lpkgs, err := packages.Load(&packages.Config{Mode: packages.LoadTypes}, pkgNames...)
	if err != nil {
		log.Fatalf("Error loading packages: %s", err)
	}

	var (
		s = time.Now()
		b bytes.Buffer
		i int
		n int
		m int
	)

	b.WriteString("// Code generated by go generate: DO NOT EDIT\n\n")
	b.WriteString("package gentests\n\n")
	b.WriteString("func init() {\n")
	b.WriteString("apiRegistry = map[string]map[string]string{\n\n")

	for _, lpkg := range lpkgs {
		n++

		log.Println(lpkg.Types)
		log.Println(strings.Repeat("-", 80))

		scope := lpkg.Types.Scope()
		for _, name := range scope.Names() {
			m++

			obj := scope.Lookup(name)

			// Skip unexported objects
			if !obj.Exported() {
				continue
			}

			// Skip non-structs
			structObj, ok := obj.Type().Underlying().(*types.Struct)
			if !ok {
				continue
			}

			// Skip non-request objects
			if !strings.HasSuffix(obj.Name(), "Request") {
				continue
			}

			i++
			log.Printf("%-3d | %s{}\n", i, obj.Name())
			b.WriteString(fmt.Sprintf("%q: map[string]string{\n", obj.Name()))

			for j := 0; j < structObj.NumFields(); j++ {
				field := structObj.Field(j)
				if debug {
					log.Printf("        %s %s", field.Name(), field.Type())
				}
				b.WriteString(fmt.Sprintf("%q: %q,\n", field.Name(), field.Type()))
			}
			if debug {
				log.Printf("\n")
			}
			b.WriteString("},\n\n")
		}
	}

	b.WriteString("}\n")
	b.WriteString("}\n")

	out, err := format.Source(b.Bytes())
	if err != nil {
		log.Println(strings.Repeat("~", 80))
		b.WriteTo(os.Stdout)
		log.Println(strings.Repeat("~", 80))
		log.Fatalf("Error formatting the source: %s", err)
	}

	outFile, err := os.Create("api_registry.gen.go")
	if err != nil {
		log.Fatalf("Error creating registry file: %s", err)
	}

	_, err = outFile.Write(out)
	if err != nil {
		log.Fatalf("Error writing output to file: %s", err)
	}

	log.Println(strings.Repeat("=", 80))
	log.Printf("Processed %d package(s) and %d object(s) in %s.\n", n, m, time.Since(s).Truncate(time.Millisecond))
}
