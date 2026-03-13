//go:build web

package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type statusResponse struct {
	DNS    dnsStatus    `json:"dns"`
	CPU    cpuStatus    `json:"cpu"`
	Memory memoryStatus `json:"memory"`
}

type dnsStatus struct {
	Running bool `json:"running"`
}

type cpuStatus struct {
	UsedPercent float64 `json:"used_percent"`
}

type memoryStatus struct {
	UsedPercent float64 `json:"used_percent"`
}

func (s *Server) handleStatus(c *fiber.Ctx) error {
	resp := statusResponse{
		DNS:    s.probeDNS(),
		CPU:    getCPUUsage(),
		Memory: getMemoryUsage(),
	}
	return c.JSON(resp)
}

func (s *Server) probeDNS() dnsStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return dnsStatus{Running: s.engine != nil && s.engine.Running()}
}

func getCPUUsage() cpuStatus {
	pcts, err := cpu.Percent(0, false)
	if err != nil || len(pcts) == 0 {
		return cpuStatus{}
	}
	pct := float64(int(pcts[0]*10)) / 10
	return cpuStatus{UsedPercent: pct}
}

func getMemoryUsage() memoryStatus {
	v, err := mem.VirtualMemory()
	if err != nil {
		return memoryStatus{}
	}
	pct := float64(int(v.UsedPercent*10)) / 10
	return memoryStatus{UsedPercent: pct}
}
