# gonode

`gonode` is a tool to help you write Go code that easily translates to Node.js addons and bindings, complete with TypeScript definitions.

## How it works

Node.js addons are written in C++ and are compiled into a shared library.  
Go has support for C FFI (Foreign Function Interface) which allows you to call C++ functions from Go, or in this case export Go functions to C++.  

`go build` is used to compile the Go code into a shared library, which then needs some glue code to be able to call these functions from Node.js (i.e.,
initialize the module, export the functions, etc.).  

The steps are as follows:

- Export functions from Go using Go's `//export` comment and add a `go generate` directive which calls `gonode`:
  ```go
	package calculator

	//go:generate gonode -dir .

	//export Sum
	func Sum(x, y float64) float64 {
		return x + y
	}
  ```

- Compile the Go code into a shared library using `go build`:
  ```bash
  $ go build -buildmode c-archive -o build/calculator.a calculator.go
  ```

- Generate the glue code using `gonode`:
	```bash
	go generate ./...
	```

- In your Node.js project, generate bindings using `node-gyp` and import the TypeScript definitions that `gonode` generated:
	```bash
	node-gyp configure
	node-gyp build
	```

	```typescript
	import { sum } from "../build/Release/calculator";

	console.log(sum(1, 2));
	```

> See the [examples](./examples) directory for a complete example.

## Usage

TODO


