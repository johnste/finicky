package main

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"

	"github.com/dop251/goja"
)

func GetConsoleMap(namespace string) map[string]interface{} {
	logFunction := func(prefix string) func(call goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			var args []string
			args = append(args, prefix)
			for _, arg := range call.Arguments {
				if arg.ExportType().Kind() == reflect.Map || arg.ExportType().Kind() == reflect.Struct || arg.ExportType().Kind() == reflect.Ptr {
					jsonString, err := json.MarshalIndent(arg.Export(), "", "  ")
					if err != nil {
						args = append(args, arg.String())
					} else {
						args = append(args, string(jsonString))
					}
				} else {
					args = append(args, arg.String())
				}
			}
			log.Println(strings.Join(args, " "))
			return goja.Undefined()
		}
	}

	return map[string]interface{}{
		"log":   logFunction("[" + namespace + "]"),
		"error": logFunction("[" + namespace + "] [error]"),
		"warn":  logFunction("[" + namespace + "] [warn]"),
	}
}
