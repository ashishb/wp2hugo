package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed version.txt
var _version string

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of HugoManager",
		Long:  `All software has versions. This is HugoManager's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Hugo version manager - %s\n", _version)
		},
	}

	rootCmd.AddCommand(versionCmd)
}
