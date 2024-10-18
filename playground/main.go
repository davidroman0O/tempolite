package main

import (
	"fmt"
	"reflect"
)

// Example function with parameters
func exampleCallback(ctx string, a int, b int) string {
	return fmt.Sprintf("ctx: %s, a: %d, b: %d", ctx, a, b)
}

func inspectFunction(fn interface{}) {
	fnType := reflect.TypeOf(fn)

	if fnType.Kind() != reflect.Func {
		// fmt.Println("Provided value is not a function")
		return
	}

	fmt.Printf("Function type: %s\n", fnType.String())

	// For non-anonymous functions, reflect can give the package path.
	fmt.Printf("Package path: %s\n", fnType.PkgPath())

	// Get a pointer to the function (address in memory).
	fnPointer := reflect.ValueOf(fn).Pointer()
	fmt.Printf("Function pointer: %x\n", fnPointer)
}

func main() {
	inspectFunction(exampleCallback)
	inspectFunction(func() {
		// fmt.Println("Anonymous function 1")
	})
	inspectFunction(func() {
		// fmt.Println("Anonymous function 2")
	})
}
