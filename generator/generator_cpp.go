package generator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/guywaldman/gonode/parser"
)

const (
	ArgPrefix                 = "go_"
	GeneratedMethodNameSuffix = "_Func"
)

type CppGenerator struct {
	addonName string
}

func NewCppGenerator(addonName string) (*CppGenerator, error) {
	return &CppGenerator{
		addonName: addonName,
	}, nil
}

// Generate code for the exported functions (wraps over the Go exported functions with C FFI)
// as well as wrappers for the C++ functions to handle arguments and return values.
func (g *CppGenerator) Generate(exportedFunctions []parser.ExportedFunction) (string, error) {
	// Generated C++ code contains the following:
	// 1. Implementations of exported functions with wrappers
	//    - For each, we:
	//      - Check the number of arguments passed
	//      - Check the argument types
	//      - Move all arguments into local variables
	//      - Call the function exported from Go
	//      - Set the return value and return it
	// 2. Exports
	//    - Init function that calls the exported functions
	// 3. NODE_MODULE macro

	var code string

	// Include headers (both the generated header from `go build` and the Node.js legacy API header
	code += fmt.Sprintf(`
	#include "../%v.h"
	#include <node.h>
`, g.addonName)

	// Open the namepsace with using statements
	code += fmt.Sprintf(`
		namespace %v {
			using namespace v8;
			using namespace node;
		`, g.addonName)

	// Add implementations for exported functions
	for _, exportedFunc := range exportedFunctions {
		wrapperCode, err := g.generateCodeForFunction(exportedFunc)
		if err != nil {
			return "", err
		}
		code += wrapperCode
	}

	// Add the `Init` method which Node.js API expects
	code += "void Init(Local<Object> exports) {"
	// Inside the `Init` method, add the exported functions
	for _, exportedFunc := range exportedFunctions {
		funcJsName := strings.ToLower(exportedFunc.Name[:1]) + exportedFunc.Name[1:]
		wrapperName := exportedFunc.Name + GeneratedMethodNameSuffix
		code += fmt.Sprintf(`
					NODE_SET_METHOD(exports, "%v", %v);
				`, funcJsName, wrapperName)
	}
	// Close the `Init` method
	code += "}"

	// Add the `NODE_MODULE` macro
	code += "NODE_MODULE(NODE_GYP_MODULE_NAME, Init)"

	// Close the namespace
	code += "}"

	return code, nil
}

// Generate code for a single exported function
func (g *CppGenerator) generateCodeForFunction(exportedFunc parser.ExportedFunction) (string, error) {
	var methodCode string
	params := exportedFunc.Params

	// Generate the code for checking the number of arguments
	methodCode += fmt.Sprintf(`
		// Check the number of arguments passed.
		if (args.Length() < %v)
		{
			// Throw an Error that is passed back to JavaScript
			isolate->ThrowException(Exception::TypeError(
					String::NewFromUtf8(isolate,
																"Wrong number of arguments (expected %v)")
						.ToLocalChecked()));
			return;
		}
	`, len(params), len(params))

	// Generate the code for checking the argument types
	for i, param := range exportedFunc.Params {
		switch param.Type {
		case "string":
			methodCode += fmt.Sprintf(`
				if (!args[%v]->IsString())
				{
					isolate->ThrowException(Exception::TypeError(
							String::NewFromUtf8(isolate,
																"Wrong arguments")
								.ToLocalChecked()));
					return;
		}`, i)
		case "number":
			methodCode += fmt.Sprintf(`
				if (!args[%v]->IsNumber())
				{
					isolate->ThrowException(Exception::TypeError(
							String::NewFromUtf8(isolate,
																"Wrong arguments")
								.ToLocalChecked()));
					return;
				}`, i)
		}
	}

	// Generate the code for moving all arguments into local variables
	for i, param := range exportedFunc.Params {
		varName := ArgPrefix + param.Name // Name of the local variable
		argIndex := strconv.Itoa(i)
		switch param.Type {
		case "string":
			methodCode += fmt.Sprintf(`
				String::Utf8Value %v_s(isolate, args[%v]);
    		char* %v = *%v_s;
			`, varName, argIndex, varName, varName)
		case "number":
			methodCode += fmt.Sprintf(`
				auto %v = args[%v].As<Number>()->Value();
			`, varName, argIndex)
		}
	}

	// Generate the code for calling the Go function
	var argsList string
	for i := range exportedFunc.Params {
		argName := ArgPrefix + exportedFunc.Params[i].Name
		argsList += fmt.Sprintf("%v, ", argName)
	}
	argsList = strings.TrimSuffix(argsList, ", ")
	methodCode += "// Call the function exported from Go.\n"
	methodCode += fmt.Sprintf(`
		auto result = %v(%v);
	`, exportedFunc.Name, argsList)

	// Collect the result
	methodCode += "// Collect the result\n"
	methodCode += "auto nodeResult = "
	switch exportedFunc.ReturnType {
	case "string":
		methodCode += "String::NewFromUtf8(isolate, result, NewStringType::kNormal).ToLocalChecked()"
	case "number":
		methodCode += "Number::New(isolate, result)"

	default:
		return "", errors.New("unsupported return type")
	}
	methodCode += ";\n"
	// Set the return value
	methodCode += "args.GetReturnValue().Set(nodeResult);"

	methodName := exportedFunc.Name + GeneratedMethodNameSuffix
	methodCode = fmt.Sprintf(`
		void %s(const FunctionCallbackInfo<Value> &args) {
			Isolate *isolate = args.GetIsolate();
			HandleScope scope(isolate);

			%s
	}`, methodName, methodCode)

	return methodCode, nil
}
