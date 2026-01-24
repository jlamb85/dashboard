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

			data := map[string]interface{}{
				"Username":      username,
				"UsedGroups":    usedGroups,
				"DefinedGroups": definedGroups,
			}
			templates.ExecuteTemplate(w, "groups.html", data)
			return
		}

		// Handle POST - create/update group
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		action := r.FormValue("action")
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

	t.ExecuteTemplate(w, "groups.html", map[string]interface{}{
		"Username":      currentUser,
		"Error":         msg,
		"UsedGroups":    usedGroups,
		"DefinedGroups": cfg.Auth.Groups,
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
