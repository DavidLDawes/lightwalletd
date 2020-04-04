package cmd

import (
	"fmt"

	"github.com/asherda/lightwalletd/common"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Dispaly lightwalletd version",
	Long:  `Dispaly lightwalletd version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lightwalletd version: ", common.Version)
		fmt.Println("from commit: ", common.GitCommit)
		fmt.Println("on: ", common.BuildDate)
		fmt.Println("by: ", common.BuildUser)

	},
}
