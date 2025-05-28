package main

import (
	"fmt"
	"os"

	"gh2/cmd"
	"gh2/core"
)

func main() {
	defer func() {
		core.GetLogger().Info("exit")
		if err := core.CloseLogger(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
