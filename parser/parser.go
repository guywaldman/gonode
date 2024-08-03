package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
)

type Parser struct {
}

func New() (*Parser, error) {
	return &Parser{}, nil
}

type ParsedExportedFunctions struct {
	Functions []ExportedFunction
}

func (p *Parser) Parse(filePath string, reader io.Reader) ([]*ExportedFunction, error) {
	var exported []*ExportedFunction

	fset := token.NewFileSet()

	parsedFile, err := parser.ParseFile(fset, filePath, reader, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Construct a stack of all nodes in the AST, since we will need to get subsequent nodes
	ast.Inspect(parsedFile, func(node ast.Node) bool {
		if node == nil {
			return true
		}

		// Check if the node is a function declaration
		if node, ok := node.(*ast.FuncDecl); ok {
			doc := node.Doc
			if doc == nil {
				return true
			}

			exportedFunction, err := p.ParseFuncDecl(node)
			if err != nil {
				log.Fatal().Err(err).Msg("Error parsing function declaration")
				return false
			}
			exported = append(exported, exportedFunction)
		}
		return true
	})

	return exported, nil
}

type ExportComment struct {
	// The symbol the export comment refers to
	Symbol string
}

const (
	JavaScriptStringType = "string"
	JavaScriptNumberType = "number"
)

type ExportedFunctionParam struct {
	Name string
	Type string
}

type ExportedFunction struct {
	Name       string
	Params     []ExportedFunctionParam
	Doc        string
	ReturnType string
}

func (f ExportedFunction) String() string {
	return fmt.Sprintf("ExportedFunction{Name: %v, Params: %v, ReturnType: %v, Doc: '%v'}", f.Name, f.Params, f.ReturnType, f.Doc)
}

func (p *Parser) ParseFuncDecl(funcDecl *ast.FuncDecl) (*ExportedFunction, error) {
	doc := funcDecl.Doc

	// Parse the comments twice (simpler, though inefficient) to first see if there is an export comment
	// and then to extract the function documentation.
	var hasExportComment bool
	for _, comment := range doc.List {
		if isExportComment(comment.Text) {
			hasExportComment = true
			break
		}
	}
	if !hasExportComment {
		return nil, nil
	}

	// Extract the documentation
	var documentation string
	for _, comment := range doc.List {
		if !isExportComment(comment.Text) {
			trimmedComment := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if trimmedComment == "" {
				continue
			}
			documentation += trimmedComment + "\n"
		}
	}
	documentation = strings.TrimSuffix(documentation, "\n")

	// Extract the parameters
	params := make([]ExportedFunctionParam, 0)
	for _, fields := range funcDecl.Type.Params.List {
		if fields.Type == nil {
			continue
		}
		for _, fieldName := range fields.Names {
			fieldName := fieldName.String()
			jsParamType, err := parseGoTypeExpr(fields.Type)
			if err != nil {
				return nil, err
			}
			params = append(params, ExportedFunctionParam{
				Name: fieldName,
				Type: jsParamType,
			})
		}
	}

	// Extract the return type
	returnType := ""
	if funcDecl.Type.Results != nil {
		if len(funcDecl.Type.Results.List) > 1 {
			return nil, errors.New("an exported function can only have one return type")
		}
		goReturnType := funcDecl.Type.Results.List[0].Type
		jsReturnType, err := parseGoTypeExpr(goReturnType)
		if err != nil {
			return nil, err
		}
		returnType = jsReturnType
	}

	return &ExportedFunction{
		Name:       funcDecl.Name.String(),
		Params:     params,
		ReturnType: returnType,
		Doc:        documentation,
	}, nil
}

// Check if the comment is an export comment
func isExportComment(comment string) bool {
	return strings.HasPrefix(strings.TrimSpace(comment), "//export ")
}

func parseGoTypeExpr(expr ast.Expr) (string, error) {
	var goReturnType string
	switch v := expr.(type) {
	case *ast.Ident:
		goReturnType = v.Name
	case *ast.StarExpr:
		switch xv := v.X.(type) {
		case *ast.SelectorExpr:
			selectorType := xv.X.(*ast.Ident).Name // e.g., C if the type is "C.char"
			selectField := xv.Sel.Name             // e.g., "char" if the type is "C.char"
			goReturnType = "*" + selectorType + "." + selectField
		default:
			return "", fmt.Errorf("unsupported star expression: %v", xv)
		}
	default:
		return "", fmt.Errorf("unsupported type: %v", expr)
	}
	jsType, err := goTypeToJsType(goReturnType)
	if err != nil {
		return "", err
	}
	return jsType, nil
}

func goTypeToJsType(goReturnType string) (string, error) {
	switch goReturnType {
	case "*C.char":
		return "string", nil
	case "float64", "float32", "int", "int8", "int16", "int32", "int64":
		return "number", nil
	default:
		return "", fmt.Errorf("unsupported type: %v", goReturnType)
	}
}
