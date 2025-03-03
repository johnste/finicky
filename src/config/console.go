package config

import (
	"encoding/json"
	"log/slog"
	"reflect"
	"strings"

	"github.com/dop251/goja"
)

func GetConsoleMap() map[string]interface{} {
	logFunction := func(level string) func(call goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			var args []string
			for _, arg := range call.Arguments {
				if arg.ExportType() == nil {
					args = append(args, arg.String())
					continue
				}

				if arg.ExportType() == nil || arg.ExportType().Kind() == reflect.Map || arg.ExportType().Kind() == reflect.Struct || arg.ExportType().Kind() == reflect.Ptr {
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
			switch level {
			case "error":
				slog.Error(strings.Join(args, " "))
			case "warn":
				slog.Warn(strings.Join(args, " "))
			default:
				slog.Info(strings.Join(args, " "))
			}
			return goja.Undefined()
		}
	}

	return map[string]interface{}{
		"log":   logFunction("info"),
		"error": logFunction("error"),
		"warn":  logFunction("warn"),
	}
}