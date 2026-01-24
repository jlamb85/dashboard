package middleware

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "time"
)

var (
    sessionSecret []byte
)

const sessionCookieName = "sd_session"

// InitSession sets the HMAC secret for signing session cookies
func InitSession(secret string) {
    if secret == "" {
        // Use a deterministic but non-empty secret to avoid nil; in prod, config should set it
        secret = "change-me-session-secret"
    }
    sessionSecret = []byte(secret)
}

// SetSession sets an authenticated session cookie for the given username
func SetSession(w http.ResponseWriter, username string, ttl time.Duration) {
    exp := time.Now().Add(ttl).Unix()
    payload := fmt.Sprintf("%s|%d", username, exp)
    sig := sign(payload)
    token := payload + "|" + sig
    cookie := &http.Cookie{
        Name:     sessionCookieName,
        Value:    base64.StdEncoding.EncodeToString([]byte(token)),
        Path:     "/",
        HttpOnly: true,
        Secure:   false, // set true when behind HTTPS
        SameSite: http.SameSiteLaxMode,
    }
    http.SetCookie(w, cookie)
}

// ClearSession removes the session cookie
func ClearSession(w http.ResponseWriter) {
    cookie := &http.Cookie{
        Name:     sessionCookieName,
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Expires:  time.Unix(0, 0),
        MaxAge:   -1,
        SameSite: http.SameSiteLaxMode,
    }
    http.SetCookie(w, cookie)
}

// GetUsername returns the username from a valid session cookie
func GetUsername(r *http.Request) (string, bool) {
    c, err := r.Cookie(sessionCookieName)
    if err != nil || c.Value == "" {
        return "", false
    }
    raw, err := base64.StdEncoding.DecodeString(c.Value)
    if err != nil {
        return "", false
    }
    parts := strings.Split(string(raw), "|")
    if len(parts) != 3 {
        return "", false
    }
    username, expStr, sig := parts[0], parts[1], parts[2]
    payload := username + "|" + expStr
    if !hmac.Equal([]byte(sig), []byte(sign(payload))) {
        return "", false
    }
    exp, err := strconv.ParseInt(expStr, 10, 64)
    if err != nil || time.Now().Unix() > exp {
        return "", false
    }
    return username, true
}

// AuthRequired enforces authentication for all non-exempt paths
func AuthRequired(enabled bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !enabled {
                next.ServeHTTP(w, r)
                return
            }
            path := r.URL.Path
            // Allow unauthenticated access to login, health, static
            if path == "/login" || path == "/logout" || path == "/health" || strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/debug/pprof/") {
                next.ServeHTTP(w, r)
                return
            }
            if _, ok := GetUsername(r); !ok {
                // Redirect to login for any GET; 401 for others
                if r.Method == http.MethodGet {
                    http.Redirect(w, r, "/login", http.StatusFound)
                    return
                }
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func sign(s string) string {
    mac := hmac.New(sha256.New, sessionSecret)
    mac.Write([]byte(s))
    return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
