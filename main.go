package main

import (
	"github.com/ksred/ccswitch/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
