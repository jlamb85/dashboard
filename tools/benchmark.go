package main

import (
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "runtime"
    "time"
)

var endpoints = []string{
    "http://localhost:8080/",
    "http://localhost:8080/servers",
    "http://localhost:8080/vms",
    "http://localhost:8080/switches",
    "http://localhost:8080/synthetics",
}

func main() {
    for {
        for _, url := range endpoints {
            start := time.Now()
            resp, err := http.Get(url)
            elapsed := time.Since(start)
            if err != nil {
                fmt.Printf("ERROR: %s: %v\n", url, err)
                continue
            }
            resp.Body.Close()
            fmt.Printf("OK: %s - %v\n", url, elapsed)
        }
        printResourceUsage()
        time.Sleep(10 * time.Second)
    }
}

func printResourceUsage() {
    // Print Go process memory usage
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("Go Mem: Alloc = %v MiB, Sys = %v MiB\n", m.Alloc/1024/1024, m.Sys/1024/1024)

    // Print system resource usage (Linux/macOS only)
    cmd := exec.Command("ps", "-o", "pid,%cpu,%mem,comm", "-p", fmt.Sprintf("%d", os.Getpid()))
    out, err := cmd.CombinedOutput()
    if err == nil {
        fmt.Print(string(out))
    }
}