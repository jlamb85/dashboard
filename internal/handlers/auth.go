package handlers

import (
    "html/template"
    "net/http"
    "strings"

    "golang.org/x/crypto/bcrypt"
    "server-dashboard/internal/config"
    "server-dashboard/internal/middleware"
)

// LoginPageHandler renders the login form
func LoginPageHandler(templates *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // If already authenticated, redirect to dashboard
        if _, ok := middleware.GetUsername(r); ok {
            http.Redirect(w, r, "/", http.StatusFound)
            return
        }
        templates.ExecuteTemplate(w, "login.html", map[string]interface{}{})
    }
}

// LoginPostHandler handles credential verification
func LoginPostHandler(cfg *config.Config, templates *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Bad Request", http.StatusBadRequest)
            return
        }
        username := strings.TrimSpace(r.FormValue("username"))
        password := r.FormValue("password")

        // If multiple users are configured, authenticate against that list
        if len(cfg.Auth.Users) > 0 {
            var matched bool
            for _, u := range cfg.Auth.Users {
                if strings.EqualFold(u.Username, username) {
                    // If Enabled is explicitly false, deny login
                    if u.Enabled == false {
                        break
                    }
                    if u.PasswordHash != "" {
                        if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err == nil {
                            matched = true
                        }
                    } else if u.Password != "" {
                        if u.Password == password {
                            matched = true
                        }
                    }
                    break
                }
            }

            if !matched {
                templates.ExecuteTemplate(w, "login.html", map[string]interface{}{"Error": "Invalid credentials"})
                return
            }

            // Set session for 8 hours
            middleware.SetSession(w, username, 8*60*60*1e9)
            http.Redirect(w, r, "/", http.StatusFound)
            return
        }

        // Single-user mode (backward compatible)
        if username != cfg.Auth.Username {
            templates.ExecuteTemplate(w, "login.html", map[string]interface{}{"Error": "Invalid credentials"})
            return
        }

        if cfg.Auth.PasswordHash != "" {
            if err := bcrypt.CompareHashAndPassword([]byte(cfg.Auth.PasswordHash), []byte(password)); err != nil {
                templates.ExecuteTemplate(w, "login.html", map[string]interface{}{"Error": "Invalid credentials"})
                return
            }
        } else {
            if cfg.Auth.Password == "" || cfg.Auth.Password != password {
                templates.ExecuteTemplate(w, "login.html", map[string]interface{}{"Error": "Invalid credentials"})
                return
            }
        }

        middleware.SetSession(w, username, 8*60*60*1e9)
        http.Redirect(w, r, "/", http.StatusFound)
    }
}

// LogoutHandler clears the session
func LogoutHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        middleware.ClearSession(w)
        http.Redirect(w, r, "/login", http.StatusFound)
    }
}
