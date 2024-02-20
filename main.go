package main

import (
	"github.com/tphoney/plex-lookup/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
