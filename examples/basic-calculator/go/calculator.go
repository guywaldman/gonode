package main

import "C"

//go:generate gonode --dir .

// Sums up two numbers
//
//export Sum
func Sum(x, y float64) float64 {
	return x + y
}

// Subtracts two numbers
//
//export Subtract
func Subtract(x, y float64) float64 {
	return x - y
}

func main() {
}
