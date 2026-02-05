package main

import (
	"fmt"
	"github.com/JFryy/qq/cli"
	"os"
)

func main() {
	rootCmd := cli.CreateRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
