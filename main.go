package main

import (
	"github.com/fwiedmann/differ/pkg/controller"
	"github.com/fwiedmann/differ/pkg/opts"
)

func main() {
	o, err := opts.Init()
	if err != nil {
		panic(err)
	}

	c := controller.New(o)
	if err = c.Run(); err != nil {
		panic(err)
	}

}
