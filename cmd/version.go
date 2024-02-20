package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of plex-lookup",
	Long:  `All software has versions. This is plex-lookups's`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("v0.0.1")
	},
}
