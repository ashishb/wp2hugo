package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of HugoManager",
	Long:  `All software has versions. This is HugoManager's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo version manager - v1.6.0")
	},
}
