package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link current directory to a Shush project",
	Long:  `Creates a .shush/project.json file associating this repo with a specific Shush project.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔗 Linking repository to Shush...")
		// TODO: List projects from backend, select one, write to .shush/project.json
		fmt.Println("🚧 Not implemented yet! (Step 4.2)")
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
