package handlers

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"

	"golang.org/x/crypto/bcrypt"
)

// PasswordChangePageHandler renders and processes password change requests for the
// currently authenticated user. New passwords are stored as bcrypt hashes and
// plaintext passwords are removed from config.
func PasswordChangePageHandler(cfg *config.Config, templates *template.Template, configPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := middleware.GetUsername(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if r.Method == http.MethodGet {
			templates.ExecuteTemplate(w, "change-password.html", map[string]interface{}{
				"Username": username,
				"IsAdmin":  isAdminUser(cfg, username),
			})
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		current := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirm := r.FormValue("confirm_password")

		if strings.TrimSpace(newPassword) == "" || newPassword != confirm {
			templates.ExecuteTemplate(w, "change-password.html", map[string]interface{}{
				"Username": username,
				"IsAdmin":  isAdminUser(cfg, username),
				"Error":    "New passwords must match and cannot be empty",
			})
			return
		}

		if err := verifyUserPassword(cfg, username, current); err != nil {
			templates.ExecuteTemplate(w, "change-password.html", map[string]interface{}{
				"Username": username,
				"IsAdmin":  isAdminUser(cfg, username),
				"Error":    "Current password is incorrect",
			})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			templates.ExecuteTemplate(w, "change-password.html", map[string]interface{}{
				"Username": username,
				"IsAdmin":  isAdminUser(cfg, username),
				"Error":    "Unable to hash password",
			})
			return
		}

		if err := updateUserPassword(cfg, configPath, username, string(hash)); err != nil {
			templates.ExecuteTemplate(w, "change-password.html", map[string]interface{}{
				"Username": username,
				"IsAdmin":  isAdminUser(cfg, username),
				"Error":    "Failed to persist new password",
			})
			return
		}

		templates.ExecuteTemplate(w, "change-password.html", map[string]interface{}{
			"Username": username,
			"IsAdmin":  isAdminUser(cfg, username),
			"Success":  "Password updated successfully",
		})
	}
}

// verifyUserPassword checks the provided password against the current config.
func verifyUserPassword(cfg *config.Config, username, password string) error {
	// Multi-user mode
	if len(cfg.Auth.Users) > 0 {
		for _, u := range cfg.Auth.Users {
			if strings.EqualFold(u.Username, username) {
				if u.Enabled == false {
					return bcrypt.ErrMismatchedHashAndPassword
				}
				if u.PasswordHash != "" {
					return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
				}
				if u.Password != "" {
					if u.Password == password {
						return nil
					}
				}
				return bcrypt.ErrMismatchedHashAndPassword
			}
		}
		return bcrypt.ErrMismatchedHashAndPassword
	}

	// Single-user mode
	if username != cfg.Auth.Username {
		return bcrypt.ErrMismatchedHashAndPassword
	}
	if cfg.Auth.PasswordHash != "" {
		return bcrypt.CompareHashAndPassword([]byte(cfg.Auth.PasswordHash), []byte(password))
	}
	if cfg.Auth.Password != "" && cfg.Auth.Password == password {
		return nil
	}
	return bcrypt.ErrMismatchedHashAndPassword
}

// updateUserPassword updates the in-memory config and the config file on disk
// to store the provided bcrypt hash and remove any plaintext password entry.
func updateUserPassword(cfg *config.Config, configPath, username, hash string) error {
	if len(cfg.Auth.Users) > 0 {
		for i, u := range cfg.Auth.Users {
			if strings.EqualFold(u.Username, username) {
				cfg.Auth.Users[i].PasswordHash = hash
				cfg.Auth.Users[i].Password = ""
				break
			}
		}
		return persistUserPassword(configPath, username, hash, true)
	}

	// Single-user
	cfg.Auth.PasswordHash = hash
	cfg.Auth.Password = ""
	return persistUserPassword(configPath, username, hash, false)
}

// persistUserPassword performs a surgical text update to avoid clobbering
// comments in config.yaml. It rewrites the relevant user or top-level auth
// password fields with the provided bcrypt hash and removes plaintext entries.
func persistUserPassword(configPath, username, hash string, multi bool) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	if multi {
		inAuth := false
		inUser := false
		userIndent := ""
		hashWritten := false

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "auth:") {
				inAuth = true
				continue
			}
			if inAuth {
				// Exit auth block when indentation drops to 0 or a new top-level key appears
				if len(strings.TrimSpace(line)) > 0 && !strings.HasPrefix(line, " ") {
					break
				}
				if strings.HasPrefix(trimmed, "users:") {
					continue
				}
				if strings.HasPrefix(trimmed, "- username:") {
					uname := strings.Trim(strings.TrimPrefix(trimmed, "- username:"), " \"")
					inUser = strings.EqualFold(uname, username)
					if idx := strings.Index(line, "-"); idx >= 0 {
						userIndent = line[:idx] + "  "
					} else {
						userIndent = "    "
					}
					hashWritten = false
					continue
				}
				if inUser {
					key := strings.SplitN(trimmed, ":", 2)[0]
					switch key {
					case "password_hash":
						lines[i] = userIndent + "password_hash: \"" + hash + "\""
						hashWritten = true
					case "password":
						// Drop plaintext password line
						lines[i] = ""
					case "username":
						// skip
					default:
						// keep other keys as-is
					}
				}
			}
		}

		// If we didn't find an existing password_hash line for this user, insert one
		if !hashWritten {
			for idx := 0; idx < len(lines); idx++ {
				trimmed := strings.TrimSpace(lines[idx])
				if strings.HasPrefix(trimmed, "- username:") {
					uname := strings.Trim(strings.TrimPrefix(trimmed, "- username:"), " \"")
					if strings.EqualFold(uname, username) {
						indent := "    "
						if idx := strings.Index(lines[idx], "-"); idx >= 0 {
							indent = lines[idx][:idx] + "  "
						}
						insertAt := idx + 1
						newLines := append([]string{}, lines[:insertAt]...)
						newLines = append(newLines, indent+"password_hash: \""+hash+"\"")
						newLines = append(newLines, lines[insertAt:]...)
						lines = newLines
						break
					}
				}
			}
		}
	} else {
		// single-user mode: operate within auth block on password and password_hash
		inAuth := false
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "auth:") {
				inAuth = true
				continue
			}
			if inAuth {
				if len(strings.TrimSpace(line)) > 0 && !strings.HasPrefix(line, " ") {
					break
				}
				key := strings.SplitN(trimmed, ":", 2)[0]
				switch key {
				case "password_hash":
					lines[i] = "  password_hash: \"" + hash + "\""
				case "password":
					// remove plaintext
					lines[i] = "  password: \"\""
				}
			}
		}
	}

	newContent := strings.Join(lines, "\n")
	return os.WriteFile(configPath, []byte(newContent), 0644)
}
