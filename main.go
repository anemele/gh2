package main

import (
	"fmt"
	"os"

	"gh2/cmd"
	"gh2/core"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	core.GetLogger().Info("exit program")
}
