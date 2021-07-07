//go:generate go-bindata -ignore .*template-bindata\.go -pkg templates -o templates/template-bindata.go templates/...
package main

import (
	"fmt"
	"os"

	"github.com/infobloxopen/atlas-cli/atlas/commands"
)

func main() {
	commandSet := commands.GetCommandSet()
	if len(os.Args) < 2 {
		fmt.Printf("Command is required. Please choose one of %v\n", commands.GetCommandNames())
		os.Exit(1)
	}
	command, ok := commandSet[os.Args[1]]
	if !ok {
		fmt.Printf("Command \"%s\" is not valid. Please choose one of %v\n", os.Args[1], commands.GetCommandNames())
		os.Exit(1)
	}
	if err := command.GetFlagset().Parse(os.Args[2:]); err != nil {
		fmt.Println("failed to parse flags:", err)
		os.Exit(1)
	}
	if err := command.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
