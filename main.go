package main

import (
	"fmt"
	"os"

	"gh2/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
