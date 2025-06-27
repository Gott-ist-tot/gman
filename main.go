package main

import (
	"gman/cmd"
	"gman/internal/errors"
)

func main() {
	if err := cmd.Execute(); err != nil {
		errors.Exit(err)
	}
}
