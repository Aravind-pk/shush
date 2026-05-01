package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run -- [command]",
	Short: "Inject secrets and execute a child process",
	Long:  `Fetches secrets for the linked project and injects them into the child process at runtime.`,
	Args:  cobra.MinimumNArgs(1), // Requires at least one argument (the command to run)
	Run: func(cmd *cobra.Command, args []string) {
		targetCmd := args[0]
		targetArgs := args[1:]

		fmt.Printf("🤫 Fetching secrets for env: [%s]...\n", envFlag)
		fmt.Printf("🏃 Executing: %s %v\n", targetCmd, targetArgs)

		// TODO: Fetch from gRPC, spawn exec.Command
		fmt.Println("🚧 Not implemented yet! (Step 4.4)")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
