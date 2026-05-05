package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JFryy/qq/cli"
)

func main() {
	// When qq is used in a pipeline and the downstream consumer exits early
	// (e.g. `qq --stream ... | grep -q foo`), the OS sends SIGPIPE to qq the
	// next time it writes to the now-closed pipe. Go's default behaviour is to
	// let SIGPIPE kill the process, which produces exit code 141 (128+13) and
	// causes shells running with `pipefail` to report a failure even though the
	// pipeline did exactly what was asked of it.
	//
	// We catch SIGPIPE and exit 0 instead, matching the behaviour of standard
	// Unix tools (grep, sed, head, etc.). The goroutine stops qq immediately
	// rather than letting it drain the rest of its input with nowhere to write.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGPIPE)
	go func() {
		<-sigCh
		os.Exit(0)
	}()

	rootCmd := cli.CreateRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
