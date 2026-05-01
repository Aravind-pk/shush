package main

import (
	"fmt"
	"os"

	"github.com/Aravind-pk/shush/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
