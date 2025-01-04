package main

import (
	"fmt"
	"reflect"
)

func a() string {
	return "a"
}

func b() string {
	return "b"
}

func main() {
	va := reflect.ValueOf(a)

	fmt.Println(va.Addr())
}
