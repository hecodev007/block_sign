package main

import (
	"runtime"

	cmd "btmSign/bytom/cmd/bytomcli/commands"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
