package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Shush backend",
	Long:  `Triggers the browser-based OAuth flow to get a local developer token.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Commencing login sequence...")
		// TODO: Implement local web server, trigger browser, save JWT
		fmt.Println("🚧 Not implemented yet! (Step 4.2)")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
