package parser

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_Parse(t *testing.T) {
	// Arrange
	parser, err := New()
	if err != nil {
		t.Fatal(err)
	}
	goCode := `
package main

import "C"

// Sums up two numbers
//
//export Sum
func Sum(x, y float64) float64 {
	return x + y
}
`
	reader := bytes.NewReader([]byte(goCode))

	// Act
	functions, err := parser.Parse("test.go", reader)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(functions))
	exportedFunction := functions[0]
	assert.Equal(t, "Sum", exportedFunction.Name)
	assert.Equal(t, "Sums up two numbers", exportedFunction.Doc)
	assert.Equal(t, 2, len(exportedFunction.Params))
	assert.Equal(t, "x", exportedFunction.Params[0].Name)
	assert.Equal(t, JavaScriptNumberType, exportedFunction.Params[0].Type)
	assert.Equal(t, "y", exportedFunction.Params[1].Name)
	assert.Equal(t, JavaScriptNumberType, exportedFunction.Params[1].Type)
	assert.Equal(t, JavaScriptNumberType, exportedFunction.ReturnType)
}
