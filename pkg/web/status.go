//go:build web

package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/wiredlush/luna-dns/pkg/engine"
)

type statusResponse struct {
	DNS    dnsStatus            `json:"dns"`
	Stats  engine.StatsSnapshot `json:"stats"`
	System systemStatus         `json:"system"`
}

type dnsStatus struct {
	Running   bool  `json:"running"`
	UptimeSec int64 `json:"uptime_sec"`
	CacheSize int   `json:"cache_size"`
}

type systemStatus struct {
	CPU  float64 `json:"cpu"`
	RAM  float64 `json:"ram"`
	Disk float64 `json:"disk"`
}

func (s *Server) handleStatus(c *fiber.Ctx) error {
	eng := s.getEngine()
	running := eng != nil && eng.Running()
	resp := statusResponse{
		DNS:    dnsStatus{Running: running},
		System: getSystemStatus(),
	}
	if eng != nil {
		resp.Stats = eng.Stats()
		if running {
			resp.DNS.UptimeSec = int64(eng.Uptime().Seconds())
			resp.DNS.CacheSize = eng.CacheSize()
		}
	}
	return c.JSON(resp)
}

func getSystemStatus() systemStatus {
	var st systemStatus

	if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
		st.CPU = float64(int(pcts[0]*10)) / 10
	}

	if v, err := mem.VirtualMemory(); err == nil {
		st.RAM = float64(int(v.UsedPercent*10)) / 10
	}

	if d, err := disk.Usage("/"); err == nil {
		st.Disk = float64(int(d.UsedPercent*10)) / 10
	}

	return st
}
