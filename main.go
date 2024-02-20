package main

import (
	"tphoney/plex-lookup/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
