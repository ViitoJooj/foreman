// Command cli is the `foreman` command-line client for the harness API.
package main

import (
	"fmt"
	"os"

	"github.com/ViitoJooj/foreman/internal/adapters/driving/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
