package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ticktick-go/internal/config"
)

// Known subcommands that should not trigger quick-add
var knownSubcommands = map[string]bool{
	"task":    true,
	"project": true,
	"auth":    true,
	"help":    true,
	"version": true,
}

var rootCmd = &cobra.Command{
	Use:   "ttg",
	Short: "TickTick CLI - Manage your tasks from the terminal",
	Long:  `A CLI tool for managing TickTick tasks with OAuth2 authentication.`,
}

var (
	jsonFlag bool
	cfg      *config.Config
)

func Execute() error {
	// Check if we should use quick-add mode
	// Look at os.Args to determine if first arg is not a known subcommand
	fmt.Fprintf(os.Stderr, "DEBUG: os.Args = %v\n", os.Args)
	if len(os.Args) > 1 {
		firstArg := os.Args[1]
		if !knownSubcommands[firstArg] && firstArg != "--help" && firstArg != "-h" {
			// Prepend "task add" to the arguments
			newArgs := append([]string{os.Args[0], "task", "add"}, os.Args[1:]...)
			fmt.Fprintf(os.Stderr, "DEBUG: newArgs = %v\n", newArgs)
			os.Args = newArgs
		}
	}
	
	cfg = config.Load()
	return rootCmd.Execute()
}

func addGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(projectCmd)
}
