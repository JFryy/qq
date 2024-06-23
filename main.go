package main

import (
	"fmt"
	"github.com/JFryy/qq/cli"
	"github.com/JFryy/qq/codec"
	"os"
)

func main() {
	_ = codec.SupportedFileTypes
	rootCmd := cli.CreateRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
