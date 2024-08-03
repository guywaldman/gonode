package generator

import (
	"fmt"
	"strings"

	"github.com/guywaldman/gonode/parser"
)

type TypeScriptGenerator struct {
	addonName string
}

func NewTypeScriptGenerator(addonName string) (*TypeScriptGenerator, error) {
	return &TypeScriptGenerator{
		addonName: addonName,
	}, nil
}

func (g *TypeScriptGenerator) Generate(exportedFunctions []parser.ExportedFunction) (string, error) {
	var code string

	// Declare the interface
	addonInterfaceName := strings.ToUpper(g.addonName[:1]) + g.addonName[1:]
	code += fmt.Sprintf("export interface %v {", addonInterfaceName)
	for _, exportedFunc := range exportedFunctions {
		typescriptArgsListStr := ""
		for _, param := range exportedFunc.Params {
			typescriptArgsListStr += fmt.Sprintf("%v: %v,", param.Name, param.Type)
		}
		typescriptArgsListStr = strings.TrimSuffix(typescriptArgsListStr, ",")
		funcJsName := strings.ToLower(exportedFunc.Name[:1]) + exportedFunc.Name[1:]
		code += fmt.Sprintf(`
		/**
		%v
		**/
			%v: (%v) => %v;
		`, exportedFunc.Doc, funcJsName, typescriptArgsListStr, exportedFunc.ReturnType)
	}
	code += "\n}"

	// Add the require statement
	code += fmt.Sprintf(`
		const addon: %v = require("%v");
	`, addonInterfaceName, g.addonName)

	// Add the export statement
	code += "export default addon;"

	return code, nil
}
