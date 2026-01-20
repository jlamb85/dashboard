package services

import (
    "context"
    "time"
    "sync"
    "server-dashboard/internal/models"
)

type MonitorService struct {
    servers []models.Server
    vms     []models.VM
    mu      sync.Mutex
}

func NewMonitorService(servers []models.Server, vms []models.VM) *MonitorService {
    return &MonitorService{
        servers: servers,
        vms:     vms,
    }
}

func (m *MonitorService) CheckServerStatus(ctx context.Context) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for i := range m.servers {
        go func(server *models.Server) {
            // Perform health check logic here
            // Update server status
        }(&m.servers[i])
    }
}

func (m *MonitorService) CheckVMStatus(ctx context.Context) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for i := range m.vms {
        go func(vm *models.VM) {
            // Perform health check logic here
            // Update VM status
        }(&m.vms[i])
    }
}

func (m *MonitorService) StartMonitoring(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.CheckServerStatus(context.Background())
            m.CheckVMStatus(context.Background())
        }
    }
}