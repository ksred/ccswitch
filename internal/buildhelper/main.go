package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ksred/ccswitch/internal/ui"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: buildhelper <type> <message>")
		os.Exit(1)
	}

	msgType := os.Args[1]
	message := strings.Join(os.Args[2:], " ")

	switch msgType {
	case "info":
		ui.Info(message)
	case "success":
		ui.Success(message)
	case "error":
		ui.Error(message)
	case "warning":
		ui.Warning(message)
	case "title":
		ui.Title(message)
	default:
		fmt.Println(message)
	}
}