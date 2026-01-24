package handlers

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
)

// GroupsPageHandler displays and manages permission groups
func GroupsPageHandler(cfg *config.Config, templates *template.Template, configPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := middleware.GetUsername(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !isAdminUser(cfg, username) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if r.Method == http.MethodGet {
			// Gather all groups in use from users
			usedGroups := make(map[string]int)
			for _, u := range cfg.Auth.Users {
				for _, g := range userGroups(u) {
					usedGroups[g]++
				}
			}

			// Get defined groups from config
			definedGroups := cfg.Auth.Groups

			// Build group members map
			groupMembers := make(map[string][]string)
			for _, g := range definedGroups {
				for _, u := range cfg.Auth.Users {
					if u.Enabled && userInGroup(u, g.Name) {
						groupMembers[g.Name] = append(groupMembers[g.Name], u.Username)
					}
				}
			}

			data := map[string]interface{}{
				"Username":      username,
				"UsedGroups":    usedGroups,
				"DefinedGroups": definedGroups,
				"GroupMembers":  groupMembers,
				"AllUsers":      cfg.Auth.Users,
			}
			templates.ExecuteTemplate(w, "groups.html", data)
			return
		}

		// Handle POST - create/update group or manage membership
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		action := r.FormValue("action")

		// Handle membership updates
		if action == "update_members" {
			groupName := strings.TrimSpace(r.FormValue("group_name"))
			if groupName == "" {
				renderGroupError(templates, username, cfg, "Group name is required", w)
				return
			}

			// Get selected members
			selectedMembers := make(map[string]bool)
			for _, u := range cfg.Auth.Users {
				if r.FormValue("member_"+u.Username) == "on" {
					selectedMembers[u.Username] = true
				}
			}

			// Update user groups
			for i := range cfg.Auth.Users {
				var updatedGroups []string
				for _, g := range cfg.Auth.Users[i].Groups {
					if !strings.EqualFold(g, groupName) {
						updatedGroups = append(updatedGroups, g)
					}
				}

				if selectedMembers[cfg.Auth.Users[i].Username] {
					updatedGroups = append(updatedGroups, groupName)
				}

				cfg.Auth.Users[i].Groups = updatedGroups
			}

			if err := writeUsersToConfig(cfg, configPath); err != nil {
				renderGroupError(templates, username, cfg, "Failed to update membership: "+err.Error(), w)
				return
			}
			http.Redirect(w, r, "/account/groups", http.StatusFound)
			return
		}

		groupName := strings.TrimSpace(r.FormValue("group_name"))
		description := strings.TrimSpace(r.FormValue("description"))
		permsRaw := strings.TrimSpace(r.FormValue("permissions"))

		if action == "delete" {
			if err := deleteGroup(cfg, configPath, groupName); err != nil {
				renderGroupError(templates, username, cfg, "Failed to delete group: "+err.Error(), w)
				return
			}
			http.Redirect(w, r, "/account/groups", http.StatusFound)
			return
		}

		if groupName == "" {
			renderGroupError(templates, username, cfg, "Group name is required", w)
			return
		}

		var permissions []string
		if permsRaw != "" {
			parts := strings.Split(permsRaw, ",")
			for _, p := range parts {
				if t := strings.TrimSpace(p); t != "" {
					permissions = append(permissions, t)
				}
			}
		}

		if err := saveGroup(cfg, configPath, groupName, description, permissions); err != nil {
			renderGroupError(templates, username, cfg, "Failed to save group: "+err.Error(), w)
			return
		}

		http.Redirect(w, r, "/account/groups", http.StatusFound)
	}
}

func renderGroupError(t *template.Template, currentUser string, cfg *config.Config, msg string, w http.ResponseWriter) {
	usedGroups := make(map[string]int)
	for _, u := range cfg.Auth.Users {
		for _, g := range userGroups(u) {
			usedGroups[g]++
		}
	}

	groupMembers := make(map[string][]string)
	for _, g := range cfg.Auth.Groups {
		for _, u := range cfg.Auth.Users {
			if u.Enabled && userInGroup(u, g.Name) {
				groupMembers[g.Name] = append(groupMembers[g.Name], u.Username)
			}
		}
	}

	t.ExecuteTemplate(w, "groups.html", map[string]interface{}{
		"Username":      currentUser,
		"Error":         msg,
		"UsedGroups":    usedGroups,
		"DefinedGroups": cfg.Auth.Groups,
		"GroupMembers":  groupMembers,
		"AllUsers":      cfg.Auth.Users,
	})
}

func saveGroup(cfg *config.Config, configPath, name, description string, permissions []string) error {
	// Update in-memory config
	found := false
	for i := range cfg.Auth.Groups {
		if strings.EqualFold(cfg.Auth.Groups[i].Name, name) {
			cfg.Auth.Groups[i].Description = description
			cfg.Auth.Groups[i].Permissions = permissions
			found = true
			break
		}
	}
	if !found {
		cfg.Auth.Groups = append(cfg.Auth.Groups, config.GroupDefinition{
			Name:        name,
			Description: description,
			Permissions: permissions,
		})
	}

	// Write to config file
	return writeGroupsToConfig(cfg, configPath)
}

func deleteGroup(cfg *config.Config, configPath, name string) error {
	// Remove from in-memory config
	var updated []config.GroupDefinition
	for _, g := range cfg.Auth.Groups {
		if !strings.EqualFold(g.Name, name) {
			updated = append(updated, g)
		}
	}
	cfg.Auth.Groups = updated

	// Write to config file
	return writeGroupsToConfig(cfg, configPath)
}

func writeGroupsToConfig(cfg *config.Config, configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	var out []string
	inAuth := false
	inGroups := false
	groupsWritten := false
	indentAuth := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "auth:") {
			inAuth = true
			indentAuth = line[:strings.Index(line, "a")]
			out = append(out, line)
			continue
		}

		if inAuth && trimmed != "" && !strings.HasPrefix(line, " ") {
			// Leaving auth block - write groups if not written
			if !groupsWritten {
				out = append(out, formatGroupsBlock(cfg.Auth.Groups, indentAuth+"  ")...)
				groupsWritten = true
			}
			inAuth = false
			inGroups = false
		}

		if inAuth && strings.HasPrefix(trimmed, "groups:") {
			inGroups = true
			// Skip groups block - we'll rewrite it
			continue
		}

		if inGroups {
			// Skip existing group entries
			if strings.HasPrefix(line, indentAuth+"    -") || strings.HasPrefix(line, indentAuth+"      ") {
				continue
			}
			// End of groups block
			inGroups = false
			out = append(out, formatGroupsBlock(cfg.Auth.Groups, indentAuth+"  ")...)
			groupsWritten = true
		}

		out = append(out, line)

		// If at end and still in auth, write groups
		if i == len(lines)-1 && inAuth && !groupsWritten {
			out = append(out, formatGroupsBlock(cfg.Auth.Groups, indentAuth+"  ")...)
			groupsWritten = true
		}
	}

	return os.WriteFile(configPath, []byte(strings.Join(out, "\n")), 0644)
}

func formatGroupsBlock(groups []config.GroupDefinition, indent string) []string {
	if len(groups) == 0 {
		return nil
	}

	var lines []string
	lines = append(lines, indent+"groups:")
	for _, g := range groups {
		lines = append(lines, indent+"  - name: \""+g.Name+"\"")
		if g.Description != "" {
			lines = append(lines, indent+"    description: \""+g.Description+"\"")
		}
		if len(g.Permissions) > 0 {
			lines = append(lines, indent+"    permissions: [\""+strings.Join(g.Permissions, "\", \"")+"\"]")
		}
	}
	return lines
}

// userInGroup checks if a user belongs to a specific group
func userInGroup(u config.UserCredential, groupName string) bool {
	for _, g := range userGroups(u) {
		if strings.EqualFold(g, groupName) {
			return true
		}
	}
	return false
}

// writeUsersToConfig updates user group membership in the config file
func writeUsersToConfig(cfg *config.Config, configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	var out []string
	inAuth := false
	indentAuth := ""
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "auth:") {
			inAuth = true
			indentAuth = line[:strings.Index(line, "a")]
			out = append(out, line)
			i++
			continue
		}

		if inAuth && trimmed != "" && !strings.HasPrefix(line, " ") {
			inAuth = false
		}

		if inAuth && strings.HasPrefix(trimmed, "users:") {
			out = append(out, line)
			i++

			// Process users section
			for i < len(lines) {
				line := lines[i]
				trimmed := strings.TrimSpace(line)

				if trimmed == "" || (!strings.HasPrefix(trimmed, "- username:") && !strings.HasPrefix(line, indentAuth+"  ")) {
					// End of users section
					break
				}

				if strings.HasPrefix(trimmed, "- username:") {
					usernamePart := strings.TrimPrefix(trimmed, "- username:")
					usernameVal := strings.Trim(strings.TrimSpace(usernamePart), "\"'")

					out = append(out, line)
					i++

					// Find this user in config
					var targetUser *config.UserCredential
					for j := range cfg.Auth.Users {
						if cfg.Auth.Users[j].Username == usernameVal {
							targetUser = &cfg.Auth.Users[j]
							break
						}
					}

					// Copy user properties
					hasGroups := false
					for i < len(lines) {
						line := lines[i]
						trimmed := strings.TrimSpace(line)

						if strings.HasPrefix(trimmed, "- username:") || (trimmed != "" && !strings.HasPrefix(line, indentAuth+"    ")) {
							// Next user or end
							if targetUser != nil && len(targetUser.Groups) > 0 && !hasGroups {
								var quoted []string
								for _, g := range targetUser.Groups {
									quoted = append(quoted, "\""+strings.TrimSpace(g)+"\"")
								}
								out = append(out, indentAuth+"    groups: ["+strings.Join(quoted, ", ")+"]")
							}
							break
						}

						if strings.HasPrefix(trimmed, "groups:") {
							hasGroups = true
							if targetUser != nil && len(targetUser.Groups) > 0 {
								var quoted []string
								for _, g := range targetUser.Groups {
									quoted = append(quoted, "\""+strings.TrimSpace(g)+"\"")
								}
								out = append(out, indentAuth+"    groups: ["+strings.Join(quoted, ", ")+"]")
							} else {
								// Skip this line - user has no groups
							}
							i++
							continue
						}

						out = append(out, line)
						i++
					}
					continue
				}

				out = append(out, line)
				i++
			}
			continue
		}

		out = append(out, line)
		i++
	}

	return os.WriteFile(configPath, []byte(strings.Join(out, "\n")), 0644)
}
