package services

import (
    "context"
    "math/rand"
    "net"
    "net/http"
    "strconv"
    "strings"
    "sync"
    "time"

    "server-dashboard/internal/config"
    "server-dashboard/internal/models"
)

var (
    syntheticResults map[string]models.SyntheticCheckResult
    syntheticMu      sync.RWMutex
)

// InitSynthetic initializes synthetic check runners based on config.
func InitSynthetic(cfg *config.Config) {
    syntheticResults = make(map[string]models.SyntheticCheckResult)
    rand.Seed(time.Now().UnixNano())

    for _, check := range cfg.SyntheticChecks {
        if !check.Enabled {
            continue
        }

        interval := time.Duration(check.IntervalSeconds) * time.Second
        if interval <= 0 {
            interval = 60 * time.Second
        }

        // Run once immediately, then on interval
        runSyntheticCheck(check, cfg.Monitoring.UseMockData)
        go runSyntheticLoop(check, interval, cfg.Monitoring.UseMockData)
    }
}

func runSyntheticLoop(check config.SyntheticCheckConfig, interval time.Duration, useMock bool) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            runSyntheticCheck(check, useMock)
        }
    }
}

func runSyntheticCheck(check config.SyntheticCheckConfig, useMock bool) {
    var result models.SyntheticCheckResult
    result.ID = check.ID
    result.Name = check.Name
    result.Type = strings.ToLower(check.Type)
    result.Tags = append([]string{}, check.Tags...)
    result.LastRun = time.Now()

    timeout := time.Duration(check.TimeoutSeconds) * time.Second
    if timeout <= 0 {
        timeout = 5 * time.Second
    }

    if useMock {
        result.Status = mockStatus()
        result.LatencyMs = int64(50 + rand.Intn(450))
        result.Message = "mocked"
        result.Target = mockTarget(check)
        saveSyntheticResult(result)
        return
    }

    switch result.Type {
    case "http":
        target := check.URL
        result.Target = target
        start := time.Now()
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        defer cancel()
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
        if err != nil {
            result.Status = "fail"
            result.Message = err.Error()
            break
        }
        client := &http.Client{Timeout: timeout}
        resp, err := client.Do(req)
        if err != nil {
            result.Status = "fail"
            result.Message = err.Error()
            break
        }
        resp.Body.Close()
        result.LatencyMs = int64(time.Since(start).Milliseconds())
        expected := check.ExpectedStatus
        if expected == 0 {
            expected = http.StatusOK
        }
        if resp.StatusCode == expected {
            result.Status = "ok"
            result.Message = "status ok"
        } else {
            result.Status = "fail"
            result.Message = "status " + http.StatusText(resp.StatusCode)
        }
    case "tcp":
        target := net.JoinHostPort(check.Host, strconv.Itoa(check.Port))
        result.Target = target
        start := time.Now()
        conn, err := net.DialTimeout("tcp", target, timeout)
        if err != nil {
            result.Status = "fail"
            result.Message = err.Error()
            break
        }
        conn.Close()
        result.LatencyMs = int64(time.Since(start).Milliseconds())
        result.Status = "ok"
        result.Message = "connect ok"
    case "dns":
        target := check.Host
        result.Target = target
        start := time.Now()
        _, err := net.DefaultResolver.LookupHost(context.Background(), target)
        if err != nil {
            result.Status = "fail"
            result.Message = err.Error()
            break
        }
        result.LatencyMs = int64(time.Since(start).Milliseconds())
        result.Status = "ok"
        result.Message = "lookup ok"
    default:
        result.Status = "fail"
        result.Message = "unknown type"
        result.Target = check.URL
    }

    saveSyntheticResult(result)
}

func mockStatus() string {
    if rand.Intn(100) < 90 {
        return "ok"
    }
    return "fail"
}

func mockTarget(check config.SyntheticCheckConfig) string {
    switch strings.ToLower(check.Type) {
    case "http":
        if check.URL != "" {
            return check.URL
        }
    case "tcp":
        if check.Host != "" && check.Port != 0 {
            return net.JoinHostPort(check.Host, strconv.Itoa(check.Port))
        }
    case "dns":
        return check.Host
    }
    return check.URL
}

func saveSyntheticResult(res models.SyntheticCheckResult) {
    syntheticMu.Lock()
    defer syntheticMu.Unlock()
    syntheticResults[res.ID] = res
}

// GetSyntheticResults returns a copy of the latest results.
func GetSyntheticResults() []models.SyntheticCheckResult {
    syntheticMu.RLock()
    defer syntheticMu.RUnlock()
    out := make([]models.SyntheticCheckResult, 0, len(syntheticResults))
    for _, v := range syntheticResults {
        out = append(out, v)
    }
    return out
}

