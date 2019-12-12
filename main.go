package main

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func valueToString(value cty.Value) string {
	if value.Type().FriendlyName() == "number" {
		var val int64
		err := gocty.FromCtyValue(value, &val)
		if err != nil {
			log.Fatalf("Failed to convert: %v", value)
		}
		return strconv.FormatInt(val, 10)
	} else if value.Type().FriendlyName() == "string" {
		var val string
		err := gocty.FromCtyValue(value, &val)
		if err != nil {
			log.Fatalf("Failed to convert: %v", value)
		}
		return val
	} else {
		log.Fatalf("Unknown type: %v", value.Type().FriendlyName())
	}

	return ""
}

func main() {
	argsWithoutProg := os.Args[1:]
	terraformDir := argsWithoutProg[0]
	variableType := argsWithoutProg[1]
	variableName := argsWithoutProg[2]

	if len(argsWithoutProg) != 3 {
		log.Fatal("usage: terraparse <terraform-directory> <local|variable> <name-of-local-or-variable>")
	}
	if variableType != "local" && variableType != "variable" {
		log.Fatal("usage: terraparse <terraform-directory> <local|variable> <name-of-local-or-variable>")
	}

	parser := hclparse.NewParser()

	files, err := ioutil.ReadDir(terraformDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tf") {
			file, diagnostics := parser.ParseHCLFile(f.Name())
			if diagnostics != nil {
				log.Fatalf("Failed to parse %v", f.Name())
			}

			bodyContent, _ := file.Body.Content(rootSchema)

			if variableType == "local" {
				for _, block := range bodyContent.Blocks {
					switch block.Type {
					case "locals":
						defs := decodeLocalsBlock(block)
						for _, local := range defs {
							if local.Name == variableName {
								value, _ := local.Expr.Value(nil)
								fmt.Println(valueToString(value))
							}
						}
					}
				}
			} else {
				for _, block := range bodyContent.Blocks {
					switch block.Type {
					case "variable":
						cfg := decodeVariableBlock(block, false)
						if cfg.Name == variableName {
							fmt.Println(valueToString(cfg.Default))
						}
					}
				}
			}
		}
	}
}
