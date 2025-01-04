package main

import (
	"fmt"

	"github.com/kijudev/go-event-bus/typeid"
)

func a() string {
	return "a"
}

func b() string {
	return "b"
}

func main() {
	g := typeid.NewFuncGenerator()

	fmt.Println("a: ", g.MustGenID(a), ", b: ", g.MustGenID(b), ", a: ", g.MustGenID(a))
}
