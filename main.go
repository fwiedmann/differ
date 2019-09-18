package main

import (
	"github.com/fwiedmann/differ/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
