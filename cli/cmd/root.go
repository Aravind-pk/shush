package cmd

import (
	"github.com/spf13/cobra"
)

var envFlag string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "shush",
	Short: "Shush is developer secrets manager",
	Long: `Shush securely manages and delivers secrets for your applications.
It allows you to securely fetch secrets from your Shush server and optionally
inject them directly into your application's runtime environment,
bypassing the need for local .env files.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flag available to all subcommands
	rootCmd.PersistentFlags().StringVarP(&envFlag, "env", "e", "", "override the environment (e.g., dev, prod)")
}
