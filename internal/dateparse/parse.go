package dateparse

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ParseDate parses natural language date strings
func ParseDate(input string, timezone string) (time.Time, error) {
	if input == "" {
		return time.Time{}, nil
	}

	input = strings.TrimSpace(input)

	// Try ISO formats first
	formats := []string{
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, f := range formats {
		if tm, err := time.Parse(f, input); err == nil {
			return tm, nil
		}
	}

	// Get timezone location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.Local
	}

	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	// Parse natural language
	lower := strings.ToLower(input)

	// Handle "today", "tomorrow", "yesterday"
	if lower == "today" {
		return today, nil
	}
	if lower == "tomorrow" {
		return today.Add(24 * time.Hour), nil
	}
	if lower == "yesterday" {
		return today.Add(-24 * time.Hour), nil
	}

	// Handle "next monday", "next friday", etc.
	weekdayMap := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}

	if strings.HasPrefix(lower, "next ") {
		dayStr := strings.TrimPrefix(lower, "next ")
		if day, ok := weekdayMap[dayStr]; ok {
			daysUntil := int(day - now.Weekday())
			if daysUntil <= 0 {
				daysUntil += 7
			}
			return today.AddDate(0, 0, daysUntil), nil
		}
	}

	// Handle "next week"
	if lower == "next week" {
		return today.AddDate(0, 0, 7), nil
	}

	// Handle "in X days/weeks/hours"
	re := regexp.MustCompile(`^in (\d+) (day|hour|minute|week)s?$`)
	matches := re.FindStringSubmatch(lower)
	if len(matches) == 3 {
		amount := 1
		fmt.Sscanf(matches[1], "%d", &amount)

		switch matches[2] {
		case "minute":
			return now.Add(time.Duration(amount) * time.Minute), nil
		case "hour":
			return now.Add(time.Duration(amount) * time.Hour), nil
		case "day":
			return today.AddDate(0, 0, amount), nil
		case "week":
			return today.AddDate(0, 0, amount*7), nil
		}
	}

	// Handle time only (3pm, 3:30pm, 15:00)
	timeRe := regexp.MustCompile(`^(\d{1,2})(?::(\d{2}))?\s*(am|pm)?$`)
	timeMatches := timeRe.FindStringSubmatch(lower)
	if len(timeMatches) > 0 {
		hour := 0
		fmt.Sscanf(timeMatches[1], "%d", &hour)
		
		minute := 0
		if len(timeMatches) > 2 && timeMatches[2] != "" {
			fmt.Sscanf(timeMatches[2], "%d", &minute)
		}
		
		// Handle am/pm
		if len(timeMatches) > 3 && timeMatches[3] != "" {
			ampm := strings.ToLower(timeMatches[3])
			if ampm == "pm" && hour < 12 {
				hour += 12
			} else if ampm == "am" && hour == 12 {
				hour = 0
			}
		}

		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc), nil
	}

	// Handle "tomorrow 3pm", "tomorrow 9am"
	comboRe := regexp.MustCompile(`^(tomorrow|today) (\d{1,2})(?::(\d{2}))?\s*(am|pm)?$`)
	comboMatches := comboRe.FindStringSubmatch(lower)
	if len(comboMatches) > 1 {
		dayStr := comboMatches[1]
		hour := 0
		fmt.Sscanf(comboMatches[2], "%d", &hour)
		
		minute := 0
		if len(comboMatches) > 3 && comboMatches[3] != "" {
			fmt.Sscanf(comboMatches[3], "%d", &minute)
		}
		
		if len(comboMatches) > 4 && comboMatches[4] != "" {
			ampm := strings.ToLower(comboMatches[4])
			if ampm == "pm" && hour < 12 {
				hour += 12
			} else if ampm == "am" && hour == 12 {
				hour = 0
			}
		}

		baseDay := today
		if dayStr == "tomorrow" {
			baseDay = today.Add(24 * time.Hour)
		}

		return time.Date(baseDay.Year(), baseDay.Month(), baseDay.Day(), hour, minute, 0, 0, loc), nil
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", input)
}
