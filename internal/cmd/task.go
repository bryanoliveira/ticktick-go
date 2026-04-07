package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"ticktick-go/internal/api"
	"ticktick-go/internal/config"
	"ticktick-go/internal/format"
)

func init() {
	taskCmd.AddCommand(taskListCmd, taskAddCmd, taskGetCmd, taskDoneCmd, taskDeleteCmd, taskEditCmd, taskItemsCmd, taskItemAddCmd, taskItemDoneCmd, taskItemDeleteCmd, taskSearchCmd)
	
	// Add global json flag to task commands
	taskListCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
	taskGetCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output in JSON format")
	
	// List flags
	taskListCmd.Flags().StringP("project", "p", "", "Filter by project name")
	taskListCmd.Flags().Bool("all", false, "Show all tasks across all projects")
	taskListCmd.Flags().String("due", "", "Filter by due date — comma-separated: today, overdue, tomorrow (e.g. --due today,overdue)")
	taskListCmd.Flags().String("priority", "", "Filter by priority — comma-separated: high, medium, low, none (e.g. --priority high,medium)")
	taskListCmd.Flags().String("tag", "", "Filter by tag")
	taskListCmd.Flags().Bool("completed", false, "Show completed tasks")
	taskListCmd.Flags().Bool("no-cache", false, "Bypass cache and fetch fresh from API")
	
	// Add flags
	taskAddCmd.Flags().StringP("project", "p", "inbox", "Project name")
	taskAddCmd.Flags().StringP("priority", "P", "", "Priority (high, medium, low)")
	taskAddCmd.Flags().StringP("due", "d", "", "Due date (natural language)")
	taskAddCmd.Flags().String("start", "", "Start date (natural language)")
	taskAddCmd.Flags().String("repeat", "", "Repeat (daily, weekly, monthly, yearly, or rrule)")
	taskAddCmd.Flags().String("tag", "", "Tags (comma-separated)")
	taskAddCmd.Flags().StringP("note", "n", "", "Task notes")
	taskAddCmd.Flags().StringP("remind", "r", "", "Reminders (comma-separated: 15m, 1h, 1d, on-time)")
	taskAddCmd.Flags().Bool("checklist", false, "Create as checklist task")
	taskAddCmd.Flags().String("items", "", "Initial checklist items (comma-separated)")
	
	// Shorthand flags for quick-add
	taskAddCmd.Flags().Bool("high", false, "High priority (shorthand for --priority high)")
	taskAddCmd.Flags().Bool("medium", false, "Medium priority (shorthand for --priority medium)")
	taskAddCmd.Flags().Bool("med", false, "Medium priority (shorthand for --priority medium)")
	taskAddCmd.Flags().Bool("low", false, "Low priority (shorthand for --priority low)")
	taskAddCmd.Flags().Bool("today", false, "Due today (shorthand for --due today)")
	taskAddCmd.Flags().Bool("tmrw", false, "Due tomorrow (shorthand for --due tomorrow)")
	taskAddCmd.Flags().Bool("tomorrow", false, "Due tomorrow (shorthand for --due tomorrow)")
	
	// Edit flags
	taskEditCmd.Flags().String("title", "", "New title")
	taskEditCmd.Flags().String("due", "", "New due date")
	taskEditCmd.Flags().String("start", "", "New start date")
	taskEditCmd.Flags().String("repeat", "", "New repeat (daily, weekly, monthly, yearly, or rrule)")
	taskEditCmd.Flags().String("priority", "", "New priority")
	taskEditCmd.Flags().String("tag", "", "New tags (comma-separated)")
	taskEditCmd.Flags().StringP("remind", "r", "", "Reminders (comma-separated: 15m, 1h, 1d, on-time)")
	taskEditCmd.Flags().String("kind", "", "New kind (checklist)")
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task management commands",
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		showAll, _ := cmd.Flags().GetBool("all")
		projectName, _ := cmd.Flags().GetString("project")
		dueFilter, _ := cmd.Flags().GetString("due")
		priorityFilter, _ := cmd.Flags().GetString("priority")
		tagFilter, _ := cmd.Flags().GetString("tag")
		showCompleted, _ := cmd.Flags().GetBool("completed")
		noCache, _ := cmd.Flags().GetBool("no-cache")

		var tasks []api.Task
		var err error

		// Any filter that needs all tasks (due filter, priority-only, --all)
		// always goes through GetAllTasksCached to benefit from the 2-min cache.
		needsAll := showAll || dueFilter != "" || (priorityFilter != "" && projectName == "")

		if needsAll {
			tasks, err = client.GetAllTasksCached(noCache)
		} else if projectName != "" {
			projectID, err := client.GetProjectIDByName(projectName)
			if err != nil {
				return err
			}
			tasks, err = client.GetProjectTasks(projectID)
		} else {
			tasks, err = client.GetInboxTasks()
		}

		if err != nil {
			return err
		}

		// Apply filters
		if dueFilter != "" {
			tasks = filterByDueMulti(tasks, dueFilter)
		}
		if priorityFilter != "" {
			tasks = filterByPriorityMulti(tasks, priorityFilter)
		}
		if tagFilter != "" {
			tasks = filterByTag(tasks, tagFilter)
		}
		if showCompleted {
			tasks = filterCompleted(tasks)
		}

		if jsonFlag {
			return format.OutputJSON(tasks)
		}

		return format.OutputTaskList(tasks, client)
	},
}

var taskAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		title := args[0]
		projectName, _ := cmd.Flags().GetString("project")
		priorityStr, _ := cmd.Flags().GetString("priority")
		dueStr, _ := cmd.Flags().GetString("due")
		startStr, _ := cmd.Flags().GetString("start")
		repeatStr, _ := cmd.Flags().GetString("repeat")
		tagsStr, _ := cmd.Flags().GetString("tag")
		note, _ := cmd.Flags().GetString("note")
		remindStr, _ := cmd.Flags().GetString("remind")
		itemsStr, _ := cmd.Flags().GetString("items")
		isChecklist, _ := cmd.Flags().GetBool("checklist")

		// Parse shorthand priority flags
		isHigh, _ := cmd.Flags().GetBool("high")
		isMedium, _ := cmd.Flags().GetBool("medium")
		isMed, _ := cmd.Flags().GetBool("med")
		isLow, _ := cmd.Flags().GetBool("low")
		if isHigh {
			priorityStr = "high"
		} else if isMedium || isMed {
			priorityStr = "medium"
		} else if isLow {
			priorityStr = "low"
		}

		// Parse shorthand due date flags
		isToday, _ := cmd.Flags().GetBool("today")
		isTmrw, _ := cmd.Flags().GetBool("tmrw")
		isTomorrow, _ := cmd.Flags().GetBool("tomorrow")
		if isToday {
			dueStr = "today"
		} else if isTmrw || isTomorrow {
			dueStr = "tomorrow"
		}

		// Determine kind
		kindStr := ""
		if isChecklist || itemsStr != "" {
			kindStr = "CHECKLIST"
		}

		// Get project ID
		var projectID string
		var err error
		if projectName == "inbox" || projectName == "" {
			projectID, err = client.GetInboxProjectID()
		} else {
			projectID, err = client.GetProjectIDByName(projectName)
		}
		if err != nil {
			return err
		}

		// Parse due date
		dueDate, err := api.ParseDueDate(dueStr, cfg.Timezone)
		if err != nil {
			return fmt.Errorf("failed to parse due date: %w", err)
		}

		// Parse start date
		startDate, err := api.ParseDueDate(startStr, cfg.Timezone)
		if err != nil {
			return fmt.Errorf("failed to parse start date: %w", err)
		}

		// Parse repeat
		repeat, err := api.ParseRepeat(repeatStr)
		if err != nil {
			return fmt.Errorf("failed to parse repeat: %w", err)
		}

		// Parse tags
		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		// Parse reminders
		reminders, err := api.ParseReminders(remindStr)
		if err != nil {
			return fmt.Errorf("failed to parse reminders: %w", err)
		}

		// Parse checklist items
		var items []api.ChecklistItem
		if itemsStr != "" {
			itemTitles := strings.Split(itemsStr, ",")
			for i, itemTitle := range itemTitles {
				itemTitle = strings.TrimSpace(itemTitle)
				if itemTitle != "" {
					items = append(items, api.ChecklistItem{
						Title:    itemTitle,
						Status:   0,
						SortOrder: int64(i * 1000),
					})
				}
			}
		}

		task := &api.Task{
			ProjectID: projectID,
			Title:     title,
			Content:   note,
			Priority:  api.ParsePriority(priorityStr),
			DueDate:   dueDate,
			StartDate: startDate,
			Repeat: repeat,
			Tags:      tags,
			IsAllDay:  dueStr != "" && !strings.ContainsAny(dueStr, "0123456789"),
			Status:    0,
			Reminders: reminders,
			Kind:      kindStr,
			Items:     items,
		}

		created, err := client.CreateTask(task)
		if err != nil {
			return err
		}
		client.InvalidateCache()

		if jsonFlag {
			return format.OutputJSON(created)
		}

		fmt.Println("✓ Task created successfully!")
		fmt.Printf("  ID: %s\n", created.ID)
		fmt.Printf("  Title: %s\n", created.Title)
		if created.StartDate != "" {
			fmt.Printf("  📅 Start: %s\n", created.StartDate)
		}
		if created.Repeat != "" {
			fmt.Printf("  🔄 Repeat: %s\n", api.RepeatToHuman(created.Repeat))
		}
		if len(created.Reminders) > 0 {
			for _, r := range created.Reminders {
				fmt.Printf("  🔔 Reminder: %s\n", api.ReminderToHuman(r.Trigger))
			}
		}
		return nil
	},
}

var taskGetCmd = &cobra.Command{
	Use:   "get [task-id]",
	Short: "Show task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		taskID := args[0]

		// Find the task across all projects
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		var task *api.Task
		var projectID string
		for _, t := range tasks {
			if t.ID == taskID {
				task = &t
				projectID = t.ProjectID
				break
			}
		}

		if task == nil {
			return fmt.Errorf("task not found: %s", taskID)
		}

		if jsonFlag {
			return format.OutputJSON(task)
		}

		return format.OutputTaskDetail(task, projectID, client)
	},
}

var taskDoneCmd = &cobra.Command{
	Use:   "done [task-id]",
	Short: "Mark a task as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		taskID := args[0]

		// Find the task to get project ID
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		var projectID string
		for _, t := range tasks {
			if t.ID == taskID {
				projectID = t.ProjectID
				break
			}
		}

		if projectID == "" {
			return fmt.Errorf("task not found: %s", taskID)
		}

		if err := client.CompleteTask(projectID, taskID); err != nil {
			return err
		}
		client.InvalidateCache()

		fmt.Println("✓ Task marked as complete!")
		return nil
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete [task-id]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		taskID := args[0]

		// Find the task to get project ID
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		var projectID string
		for _, t := range tasks {
			if t.ID == taskID {
				projectID = t.ProjectID
				break
			}
		}

		if projectID == "" {
			return fmt.Errorf("task not found: %s", taskID)
		}

		if err := client.DeleteTask(projectID, taskID); err != nil {
			return err
		}
		client.InvalidateCache()

		fmt.Println("✓ Task deleted!")
		return nil
	},
}

var taskSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tasks by title",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		query := args[0]

		// Get all tasks
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		// Filter by search query (case-insensitive substring match)
		var filtered []api.Task
		queryLower := strings.ToLower(query)
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Title), queryLower) {
				filtered = append(filtered, t)
			}
		}

		if jsonFlag {
			return format.OutputJSON(filtered)
		}

		if len(filtered) == 0 {
			fmt.Printf("No tasks found matching \"%s\"\n", query)
			return nil
		}

		fmt.Printf("Found %d task(s) matching \"%s\":\n", len(filtered), query)
		return format.OutputTaskList(filtered, client)
	},
}

var taskEditCmd = &cobra.Command{
	Use:   "edit [task-id]",
	Short: "Edit a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		taskID := args[0]
		title, _ := cmd.Flags().GetString("title")
		dueStr, _ := cmd.Flags().GetString("due")
		startStr, _ := cmd.Flags().GetString("start")
		repeatStr, _ := cmd.Flags().GetString("repeat")
		priorityStr, _ := cmd.Flags().GetString("priority")
		tagsStr, _ := cmd.Flags().GetString("tag")
		remindStr, _ := cmd.Flags().GetString("remind")
		kindStr, _ := cmd.Flags().GetString("kind")

		// Find the task
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		var task *api.Task
		for i := range tasks {
			if tasks[i].ID == taskID {
				task = &tasks[i]
				break
			}
		}

		if task == nil {
			return fmt.Errorf("task not found: %s", taskID)
		}

		// Update fields
		if title != "" {
			task.Title = title
		}
		if dueStr != "" {
			dueDate, err := api.ParseDueDate(dueStr, cfg.Timezone)
			if err != nil {
				return err
			}
			task.DueDate = dueDate
		}
		if startStr != "" {
			startDate, err := api.ParseDueDate(startStr, cfg.Timezone)
			if err != nil {
				return err
			}
			task.StartDate = startDate
		}
		if repeatStr != "" {
			repeat, err := api.ParseRepeat(repeatStr)
			if err != nil {
				return err
			}
			task.Repeat = repeat
		}
		if priorityStr != "" {
			task.Priority = api.ParsePriority(priorityStr)
		}
		if tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			task.Tags = tags
		}
		if remindStr != "" {
			reminders, err := api.ParseReminders(remindStr)
			if err != nil {
				return fmt.Errorf("failed to parse reminders: %w", err)
			}
			task.Reminders = reminders
		}
		if kindStr != "" {
			task.Kind = kindStr
		}

		_, err = client.UpdateTask(task)
		if err != nil {
			return err
		}
		client.InvalidateCache()

		fmt.Println("✓ Task updated!")
		return nil
	},
}

// Helper functions for filtering

// filterByDueMulti supports comma-separated due values: today, overdue, tomorrow
func filterByDueMulti(tasks []api.Task, filter string) []api.Task {
	parts := strings.Split(filter, ",")
	resultSet := make(map[string]api.Task)
	for _, part := range parts {
		for _, t := range filterByDue(tasks, strings.TrimSpace(part)) {
			resultSet[t.ID] = t
		}
	}
	var result []api.Task
	for _, t := range resultSet {
		result = append(result, t)
	}
	return result
}

// filterByDue filters tasks by a single due value: today, overdue, tomorrow
func filterByDue(tasks []api.Task, filter string) []api.Task {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	var filtered []api.Task
	for _, t := range tasks {
		// Skip completed tasks
		if t.Status == 2 {
			continue
		}

		due := api.ToLocalTime(t.DueDate)
		start := api.ToLocalTime(t.StartDate)

		switch filter {
		case "today":
			// Include if dueDate is today
			duedToday := !due.IsZero() && !due.Before(todayStart) && due.Before(todayEnd)
			// Include if startDate <= today (started, regardless of dueDate)
			// but only if not overdue (due < today) — overdue tasks have their own filter
			startedByToday := !start.IsZero() && start.Before(todayEnd) && (due.IsZero() || !due.Before(todayStart))
			if duedToday || startedByToday {
				filtered = append(filtered, t)
			}
		case "overdue":
			if !due.IsZero() && due.Before(todayStart) {
				filtered = append(filtered, t)
			}
		case "tomorrow":
			tomorrowStart := todayEnd
			tomorrowEnd := tomorrowStart.Add(24 * time.Hour)
			if !due.IsZero() && !due.Before(tomorrowStart) && due.Before(tomorrowEnd) {
				filtered = append(filtered, t)
			}
		}
	}
	return filtered
}

// filterByPriorityMulti supports comma-separated priority values: high, medium, low, none
func filterByPriorityMulti(tasks []api.Task, filter string) []api.Task {
	parts := strings.Split(filter, ",")
	wanted := make(map[int]bool)
	for _, p := range parts {
		wanted[api.ParsePriority(strings.TrimSpace(p))] = true
	}
	var filtered []api.Task
	for _, t := range tasks {
		if wanted[t.Priority] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func filterByPriority(tasks []api.Task, filter string) []api.Task {
	return filterByPriorityMulti(tasks, filter)
}


func filterByTag(tasks []api.Task, filter string) []api.Task {
	var filtered []api.Task
	for _, t := range tasks {
		for _, tag := range t.Tags {
			if strings.EqualFold(tag, filter) {
				filtered = append(filtered, t)
				break
			}
		}
	}
	return filtered
}

func filterCompleted(tasks []api.Task) []api.Task {
	var filtered []api.Task
	for _, t := range tasks {
		if t.Status == 2 {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// taskItemsCmd lists checklist items for a task
var taskItemsCmd = &cobra.Command{
	Use:   "items [task-id]",
	Short: "List checklist items for a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		taskID := args[0]

		// Find project ID
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		var projectID string
		for _, t := range tasks {
			if t.ID == taskID {
				projectID = t.ProjectID
				break
			}
		}

		if projectID == "" {
			return fmt.Errorf("task not found: %s", taskID)
		}

		items, err := client.GetChecklistItems(projectID, taskID)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			fmt.Println("No checklist items found.")
			return nil
		}

		fmt.Println()
		for _, item := range items {
			checkbox := "[ ]"
			if item.Status == 2 {
				checkbox = "[x]"
			}
			fmt.Printf("  %s %s (id: %s)\n", checkbox, item.Title, item.ID)
		}
		fmt.Println()
		return nil
	},
}

// taskItemAddCmd adds a checklist item to a task
var taskItemAddCmd = &cobra.Command{
	Use:   "item-add [task-id] [title]",
	Short: "Add a checklist item to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		taskID := args[0]
		title := args[1]

		// Find project ID
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		var projectID string
		for _, t := range tasks {
			if t.ID == taskID {
				projectID = t.ProjectID
				break
			}
		}

		if projectID == "" {
			return fmt.Errorf("task not found: %s", taskID)
		}

		created, err := client.AddChecklistItem(projectID, taskID, title)
		if err != nil {
			return err
		}

		fmt.Println("✓ Checklist item added!")
		fmt.Printf("  ID: %s\n", created.ID)
		fmt.Printf("  Title: %s\n", created.Title)
		return nil
	},
}

// taskItemDoneCmd marks a checklist item as complete
var taskItemDoneCmd = &cobra.Command{
	Use:   "item-done [task-id] [item-id]",
	Short: "Mark a checklist item as complete",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		taskID := args[0]
		itemID := args[1]

		// Find project ID
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		var projectID string
		for _, t := range tasks {
			if t.ID == taskID {
				projectID = t.ProjectID
				break
			}
		}

		if projectID == "" {
			return fmt.Errorf("task not found: %s", taskID)
		}

		// Get current item
		items, err := client.GetChecklistItems(projectID, taskID)
		if err != nil {
			return err
		}

		var item *api.ChecklistItem
		for i := range items {
			if items[i].ID == itemID {
				item = &items[i]
				break
			}
		}

		if item == nil {
			return fmt.Errorf("checklist item not found: %s", itemID)
		}

		// Update status to done (2)
		item.Status = 2

		_, err = client.UpdateChecklistItem(projectID, taskID, item)
		if err != nil {
			return err
		}

		fmt.Println("✓ Checklist item marked as complete!")
		return nil
	},
}

// taskItemDeleteCmd deletes a checklist item
var taskItemDeleteCmd = &cobra.Command{
	Use:   "item-delete [task-id] [item-id]",
	Short: "Delete a checklist item",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		client := api.NewClient(cfg)

		taskID := args[0]
		itemID := args[1]

		// Find project ID
		tasks, err := client.GetAllTasks()
		if err != nil {
			return err
		}

		var projectID string
		for _, t := range tasks {
			if t.ID == taskID {
				projectID = t.ProjectID
				break
			}
		}

		if projectID == "" {
			return fmt.Errorf("task not found: %s", taskID)
		}

		if err := client.DeleteChecklistItem(projectID, taskID, itemID); err != nil {
			return err
		}

		fmt.Println("✓ Checklist item deleted!")
		return nil
	},
}
