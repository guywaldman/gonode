package generator

import (
	"github.com/guywaldman/gonode/parser"
)

type GoNodeGenerator struct {
	addonName    string
	outputDir    string
	cppGenerator *CppGenerator
	tsGenerator  *TypeScriptGenerator
}

func New(addonName string, outputDir string) (*GoNodeGenerator, error) {
	cppGenerator, err := NewCppGenerator(addonName)
	if err != nil {
		return nil, err
	}
	tsGenerator, err := NewTypeScriptGenerator(addonName)
	if err != nil {
		return nil, err
	}
	return &GoNodeGenerator{
		addonName:    addonName,
		outputDir:    outputDir,
		cppGenerator: cppGenerator,
		tsGenerator:  tsGenerator,
	}, nil
}

type Generation struct {
	CppCode        string
	TypeScriptCode string
}

type Generator interface {
	Generate(exportedFunctions []parser.ExportedFunction) (*Generation, error)
}

func (g *GoNodeGenerator) Generate(exportedFunctions []parser.ExportedFunction) (*Generation, error) {
	cppCode, err := g.cppGenerator.Generate(exportedFunctions)
	if err != nil {
		return nil, err
	}
	typeScriptCode, err := g.tsGenerator.Generate(exportedFunctions)
	if err != nil {
		return nil, err
	}

	return &Generation{
		CppCode:        cppCode,
		TypeScriptCode: typeScriptCode,
	}, nil
}
