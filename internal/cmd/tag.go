package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"ticktick-go/internal/api"
	"ticktick-go/internal/config"
	"ticktick-go/internal/format"
)

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.AddCommand(tagListCmd)
	tagListCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
}

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Tag management commands",
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags used across tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		// Get all tasks
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		// Collect all unique tags
		tagSet := make(map[string]bool)
		for _, t := range tasks {
			for _, tag := range t.Tags {
				if tag != "" {
					tagSet[tag] = true
				}
			}
		}

		// Convert to slice and sort
		var tags []string
		for tag := range tagSet {
			tags = append(tags, tag)
		}
		sort.Strings(tags)

		if jsonFlag {
			return format.OutputJSON(tags)
		}

		if len(tags) == 0 {
			fmt.Println("No tags found.")
			return nil
		}

		fmt.Printf("Found %d tag(s):\n", len(tags))
		for _, tag := range tags {
			fmt.Printf("  %s\n", tag)
		}

		return nil
	},
}
