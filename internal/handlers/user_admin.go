package handlers

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
)

// UserCreatePageHandler lets an admin create new users with bcrypt hashes. It appends
// to config and clears any plaintext password.
func UserCreatePageHandler(cfg *config.Config, templates *template.Template, configPath string) http.HandlerFunc {
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
			templates.ExecuteTemplate(w, "new-user.html", map[string]interface{}{"Username": username})
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		newUser := strings.TrimSpace(r.FormValue("new_username"))
		password := r.FormValue("new_password")
		confirm := r.FormValue("confirm_password")
		groupsRaw := strings.TrimSpace(r.FormValue("groups"))
		if groupsRaw == "" {
			// Fallback for older forms that still post "roles"
			groupsRaw = strings.TrimSpace(r.FormValue("roles"))
		}

		if newUser == "" || password == "" {
			renderUserError(templates, username, "Username and password are required", w)
			return
		}
		if password != confirm {
			renderUserError(templates, username, "Passwords do not match", w)
			return
		}
		if userExists(cfg, newUser) {
			renderUserError(templates, username, "User already exists", w)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			renderUserError(templates, username, "Unable to hash password", w)
			return
		}

		var groups []string
		if groupsRaw != "" {
			parts := strings.Split(groupsRaw, ",")
			for _, p := range parts {
				if t := strings.TrimSpace(p); t != "" {
					groups = append(groups, t)
				}
			}
		}

		if err := appendUser(cfg, configPath, newUser, string(hash), groups); err != nil {
			renderUserError(templates, username, "Failed to save user", w)
			return
		}

		templates.ExecuteTemplate(w, "new-user.html", map[string]interface{}{
			"Username": username,
			"Success":  "User created with hashed password",
		})
	}
}

func renderUserError(t *template.Template, currentUser, msg string, w http.ResponseWriter) {
	t.ExecuteTemplate(w, "new-user.html", map[string]interface{}{
		"Username": currentUser,
		"Error":    msg,
	})
}

func userExists(cfg *config.Config, uname string) bool {
	for _, u := range cfg.Auth.Users {
		if strings.EqualFold(u.Username, uname) {
			return true
		}
	}
	if len(cfg.Auth.Users) == 0 && strings.EqualFold(cfg.Auth.Username, uname) {
		return true
	}
	return false
}

func isAdminUser(cfg *config.Config, uname string) bool {
	// Require membership in the admin group (with backward-compatibility for legacy admin user)
	for _, u := range cfg.Auth.Users {
		if strings.EqualFold(u.Username, uname) {
			if hasGroup(u, "admin") {
				return true
			}
			// Backward compatibility: if no groups/roles are set, allow built-in admin user
			if len(userGroups(u)) == 0 && strings.EqualFold(u.Username, "admin") {
				return true
			}
		}
	}
	if len(cfg.Auth.Users) == 0 && strings.EqualFold(cfg.Auth.Username, uname) {
		return true
	}
	return false
}

// appendUser updates in-memory cfg and writes to config file by appending a user
// with password_hash and removing any plaintext password.
func appendUser(cfg *config.Config, configPath, uname, hash string, groups []string) error {
	cfg.Auth.Users = append(cfg.Auth.Users, config.UserCredential{
		Username:     uname,
		PasswordHash: hash,
		Password:     "",
		Enabled:      true,
		Roles:        groups, // keep in sync for backward compatibility
		Groups:       groups,
	})

	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	// Try to find auth.users block to append into. If not found, create it.
	var out []string
	inAuth := false
	usersFound := false
	indentAuth := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "auth:") {
			inAuth = true
			indentAuth = line[:strings.Index(line, "a")]
		} else if inAuth {
			// exit auth block when reaching top-level key
			if trimmed != "" && !strings.HasPrefix(line, " ") {
				// inject users block before leaving if missing
				if !usersFound {
					out = append(out, indentAuth+"  users:")
					out = append(out, formatUserEntry(uname, hash, groups, indentAuth+"    ")...)
					usersFound = true
				}
				inAuth = false
			}
		}

		out = append(out, line)

		if inAuth {
			if strings.HasPrefix(trimmed, "users:") {
				usersFound = true
			}
			if usersFound && strings.HasPrefix(trimmed, "users:") {
				// Append after the users: line if no existing users are present
				// but only if next line is not already a user
				if i+1 < len(lines) {
					nextTrim := strings.TrimSpace(lines[i+1])
					if !strings.HasPrefix(nextTrim, "-") {
						out = append(out, formatUserEntry(uname, hash, groups, indentAuth+"    ")...)
						usersFound = true
					}
				}
			}
		}
	}

	if usersFound {
		// Users block exists; append at end of auth.users
		out = appendUserAtEnd(out, uname, hash, groups)
	} else if !usersFound && !inAuth {
		// No auth block found; append whole block
		out = append(out, "auth:")
		out = append(out, "  enabled: true")
		out = append(out, "  users:")
		out = append(out, formatUserEntry(uname, hash, groups, "    ")...)
	}

	return os.WriteFile(configPath, []byte(strings.Join(out, "\n")), 0644)
}

func formatUserEntry(uname, hash string, groups []string, indent string) []string {
	var lines []string
	lines = append(lines, indent+"- username: \""+uname+"\"")
	lines = append(lines, indent+"  password_hash: \""+hash+"\"")
	lines = append(lines, indent+"  enabled: true")
	if len(groups) > 0 {
		lines = append(lines, indent+"  groups: [\""+strings.Join(groups, "\", \"")+"\"]")
	}
	return lines
}

func appendUserAtEnd(lines []string, uname, hash string, groups []string) []string {
	// Append just before leaving auth block if possible; otherwise at end.
	var out []string
	inAuth := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "auth:") {
			inAuth = true
		} else if inAuth && trimmed != "" && !strings.HasPrefix(line, " ") {
			// leaving auth block
			indent := "    "
			out = append(out, formatUserEntry(uname, hash, groups, indent)...)
			inAuth = false
		}
		out = append(out, line)
		// If at end and still in auth, append
		if i == len(lines)-1 && inAuth {
			indent := "    "
			out = append(out, formatUserEntry(uname, hash, groups, indent)...)
			inAuth = false
		}
	}
	return out
}

// userGroups returns the union of groups and legacy roles in a case-insensitive way.
func userGroups(u config.UserCredential) []string {
	seen := make(map[string]bool)
	var out []string
	for _, g := range u.Groups {
		key := strings.ToLower(strings.TrimSpace(g))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, g)
	}
	for _, r := range u.Roles {
		key := strings.ToLower(strings.TrimSpace(r))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	return out
}

// hasGroup checks membership against groups and legacy roles.
func hasGroup(u config.UserCredential, target string) bool {
	for _, g := range userGroups(u) {
		if strings.EqualFold(g, target) {
			return true
		}
	}
	return false
}
