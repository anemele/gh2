package main

import (
	"fmt"
)

type ConfigCmd struct{}

func (c ConfigCmd) Run() error {
	return fmt.Errorf("TODO: config command")
}
