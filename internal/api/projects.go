package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Project represents a TickTick project
type Project struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Color         string      `json:"color,omitempty"`
	Archived      bool        `json:"archived,omitempty"`
	ParentID      string      `json:"parentId,omitempty"`
	Kind          interface{} `json:"kind,omitempty"` // 0=Personal, 1=Business (can be string or int)
	Share         bool        `json:"share,omitempty"`
	OwnerID       string      `json:"ownerId,omitempty"`
	GroupID       string      `json:"groupId,omitempty"`
	Inbox         bool        `json:"inbox,omitempty"` // True if this is the inbox project
	SortOrder     int         `json:"sortOrder,omitempty"`
	TaskCount     int         `json:"taskCount,omitempty"` // Not from API, computed
}

// GetProjects returns all projects
func (c *Client) GetProjects() ([]Project, error) {
	// Get projects from /project endpoint
	data, err := c.doRequest("GET", "/project", nil)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		// Try parsing as a map to see if there's nested data
		var respMap map[string]interface{}
		if json.Unmarshal(data, &respMap) == nil {
			// Check if there's a "projects" or "projectProfiles" field
			if projectsData, ok := respMap["projects"].([]interface{}); ok {
				for _, p := range projectsData {
					if pMap, ok := p.(map[string]interface{}); ok {
						pBytes, _ := json.Marshal(pMap)
						var proj Project
						if json.Unmarshal(pBytes, &proj) == nil {
							projects = append(projects, proj)
						}
					}
				}
			} else if profilesData, ok := respMap["projectProfiles"].([]interface{}); ok {
				for _, p := range profilesData {
					if pMap, ok := p.(map[string]interface{}); ok {
						pBytes, _ := json.Marshal(pMap)
						var proj Project
						if json.Unmarshal(pBytes, &proj) == nil {
							projects = append(projects, proj)
						}
					}
				}
			}
		}
		if len(projects) == 0 {
			return nil, fmt.Errorf("failed to parse projects: %v", err)
		}
	}

	// Try to also get project folders/groups and their projects
	folderData, folderErr := c.doRequest("GET", "/project/folder", nil)
	if folderErr == nil && len(folderData) > 0 {
		var folders []map[string]interface{}
		if json.Unmarshal(folderData, &folders) == nil {
			// For each folder, try to get its projects
			for _, folder := range folders {
				if folderID, ok := folder["id"].(string); ok {
					// Try to get projects in this folder
					folderProjectsData, _ := c.doRequest("GET", "/project/folder/"+folderID, nil)
					if len(folderProjectsData) > 0 {
						var folderProjects []Project
						if json.Unmarshal(folderProjectsData, &folderProjects) == nil {
							for _, fp := range folderProjects {
								// Check if already in list
								exists := false
								for _, p := range projects {
									if p.ID == fp.ID {
										exists = true
										break
									}
								}
								if !exists {
									projects = append(projects, fp)
								}
							}
						}
					}
				}
			}
		}
	}

	// Deduplicate by ID
	seen := make(map[string]bool)
	var unique []Project
	for _, p := range projects {
		if !seen[p.ID] {
			seen[p.ID] = true
			unique = append(unique, p)
		}
	}

	return unique, nil
}

// GetProject returns a single project by ID
func (c *Client) GetProject(projectID string) (*Project, error) {
	data, err := c.doRequest("GET", "/project/"+projectID, nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// GetProjectIDByName finds a project ID by name (case-insensitive)
func (c *Client) GetProjectIDByName(name string) (string, error) {
	projects, err := c.GetProjects()
	if err != nil {
		return "", err
	}

	nameLower := strings.ToLower(name)
	for _, p := range projects {
		if strings.ToLower(p.Name) == nameLower {
			return p.ID, nil
		}
	}

	return "", fmt.Errorf("project not found: %s", name)
}
